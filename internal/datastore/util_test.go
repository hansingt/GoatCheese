package datastore

import (
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
)

type baseTestSuite struct {
	suite.Suite
	storagePath string
}

func (suite *baseTestSuite) SetupTest() {
	assert := suite.Require()
	cfg := &config{
		Database: databaseConfig{
			Driver:           "sqlite3",
			ConnectionString: ":memory:",
		},
		Indexes: []indexConfig{},
	}
	// Initialize the database
	err := setupDatabase(cfg)
	assert.Nil(err)
	// Create a storage path
	suite.storagePath, err = ioutil.TempDir(os.TempDir(), "")
	assert.Nil(err)
}

func (suite *baseTestSuite) TearDownTest() {
	_ = db.Close()
	_ = os.RemoveAll(suite.storagePath)
	db = nil
}
