package datastore

import (
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
)

type TestSuiteWithDatastore struct {
	suite.Suite
	storagePath string
	db          *datastore
}

func (suite *TestSuiteWithDatastore) SetupTest() {
	assert := suite.Require()
	cfg := &config{
		Database: databaseConfig{
			Driver:           "sqlite3",
			ConnectionString: ":memory:",
		},
		Indexes: []indexConfig{},
	}
	// Initialize the database
	var err error
	suite.db, err = setupDatabase(cfg)
	assert.Nil(err)
	// Create a storage path
	suite.storagePath, err = ioutil.TempDir(os.TempDir(), "")
	assert.Nil(err)
}

func (suite *TestSuiteWithDatastore) TearDownTest() {
	suite.Require().Nil(suite.db.Close(), "unable to close the database connection")
	suite.Require().Nil(os.RemoveAll(suite.storagePath), "unable to remove the storage path")
}
