package datastore

import (
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
	"text/template"
)

type DatastoreTestSuite struct {
	suite.Suite
	storagePath       string
	configurationFile string
	configuration     *config
}

func TestDatastore(t *testing.T) {
	suite.Run(t, new(DatastoreTestSuite))
}

func (suite *DatastoreTestSuite) SetupTest() {
	require := suite.Require()
	var err error
	var configFile *os.File
	// Create the storage path
	suite.storagePath, err = ioutil.TempDir(os.TempDir(), "")
	require.Nil(err, "unable to create the temporary storage path")
	// create the configuration file
	configFile, err = ioutil.TempFile(os.TempDir(), "")
	require.Nil(err, "unable to create the temporary configuration file")
	// create the configuration
	suite.configuration = &config{
		StoragePath: suite.storagePath,
		Indexes: []indexConfig{
			{Name: "base", Bases: []string{}},
			{Name: "test", Bases: []string{"base"}},
		},
		Database: databaseConfig{
			Driver:           "sqlite3",
			ConnectionString: ":memory:",
		},
	}
	tmpl := template.Must(template.New("TestReadConfigurationFile").Parse(`
storagePath: "{{.StoragePath}}"
database:
  driver: "{{.Database.Driver}}"
  connection: "{{.Database.ConnectionString}}"
indexes:
{{- range $index := .Indexes }}
  - name: "{{$index.Name}}"
    bases: {{$index.Bases}}
{{- end }}
`))
	require.Nil(tmpl.Execute(configFile, suite.configuration), "unable to write the configuration file template")
	configFile.Close()
	suite.configurationFile = configFile.Name()
}

func (suite *DatastoreTestSuite) TearDownTest() {
	suite.Require().Nil(os.Remove(suite.configurationFile), "unable to remove the configuration file")
	suite.Require().Nil(os.RemoveAll(suite.storagePath), "unable to remove the storage path")
}

func (suite *DatastoreTestSuite) TestReadConfigurationFile() {
	require := suite.Require()

	checkConfig, err := readConfigurationFile(suite.configurationFile)
	require.Nil(err, "error reading the configuration file")

	// And check the results
	require.Equal(suite.configuration.StoragePath, checkConfig.StoragePath)
	require.Equal(suite.configuration.Database.Driver, checkConfig.Database.Driver)
	require.Equal(suite.configuration.Database.ConnectionString, checkConfig.Database.ConnectionString)
	for i, index := range suite.configuration.Indexes {
		require.Equal(index.Name, checkConfig.Indexes[i].Name)
		require.Equal(index.Bases, checkConfig.Indexes[i].Bases)
	}
}

func (suite *DatastoreTestSuite) TestSetupDatabase() {
	// Make sure the there is no old database
	if db != nil {
		suite.Require().Nil(db.Close())
		db = nil
	}
	// now setup the new database
	suite.Require().Nil(setupDatabase(suite.configuration), "unable to setup the database")
	suite.Require().NotNil(db, "no new database has been initialized")
	suite.Require().Nil(db.Close(), "unable to close the database")
	db = nil
}

func (suite *DatastoreTestSuite) TestAddAndGetRepositories() {
	assert := suite.Require()
	// Setup the database
	assert.Nil(setupDatabase(suite.configuration), "unable to setup the database")

	// initially, no repositories are in the database
	repos, err := AllRepositories()
	assert.Nil(err, "unable to get the repositories")
	assert.Equal([]IRepository{}, repos)

	// now add the repositories from the configuration
	assert.Nil(addRepositories(suite.configuration), "unable to add the repositories to the database")
	repos, err = AllRepositories()
	assert.Nil(err, "unable to get the repositories")
	assert.Equal(len(suite.configuration.Indexes), len(repos), "the number of repositories does not match")
	for _, index := range suite.configuration.Indexes {
		found := false
		for _, check := range repos {
			if index.Name == check.Name() {
				found = true
				break
			}
		}
		assert.True(found, "the repository '"+index.Name+"' not found!")
	}

	// modifications of the configuration will be added to the database on next run
	suite.configuration.Indexes = append(suite.configuration.Indexes, indexConfig{
		Name:  "test2",
		Bases: []string{},
	})
	assert.Nil(addRepositories(suite.configuration), "unable to add the repositories to the database")
	repos, err = AllRepositories()
	assert.Nil(err, "unable to get the repositories")
	assert.Equal(len(suite.configuration.Indexes), len(repos), "the number of repositories does not match")
	suite.configuration.Indexes[1].Bases = []string{suite.configuration.Indexes[0].Name, "test2"}
	assert.Nil(addRepositories(suite.configuration), "unable to add the repositories to the database")

	// Tear down the database
	assert.Nil(db.Close())
	db = nil
}

func (suite *DatastoreTestSuite) TestNew() {
	suite.Require().Nil(New(suite.configurationFile), "unable to create a new data store")
	suite.Require().NotNil(db, "no database has been initialized")
}

func (suite *DatastoreTestSuite) TestGetRepository() {
	suite.Require().Nil(New(suite.configurationFile), "unable to create a new data store")
	for _, index := range suite.configuration.Indexes {
		repo, err := GetRepository(index.Name)
		suite.Require().Nil(err, "unable to get the repository from the database")
		suite.Require().Equal(index.Name, repo.Name())
		bases, err := repo.Bases()
		suite.Require().Nil(err, "unable to get the repository bases")
		baseNames := make([]string, len(bases))
		for i, base := range bases {
			baseNames[i] = base.Name()
		}
		suite.Require().Equal(index.Bases, baseNames, "the list of base names does not equal")
	}
}
