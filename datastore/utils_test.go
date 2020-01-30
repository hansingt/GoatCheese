package datastore

import (
	"io/ioutil"
	"os"
	"testing"
)

func withDatabase(testFunc func() error) func() error {
	return func() error {
		cfg := &config{
			Database: databaseConfig{
				Driver:           "sqlite3",
				ConnectionString: ":memory:",
			},
			Indexes: []indexConfig{},
		}
		// Initialize the database
		err := setupDatabase(cfg)
		if err != nil {
			return err
		}
		//noinspection GoUnhandledErrorResult
		defer db.Close()
		return testFunc()
	}
}

func withStorage(testFunc func(storagePath string) error) func() error {
	return func() error {
		storagePath, err := ioutil.TempDir("", "")
		if err != nil {
			return err
		}
		//noinspection GoUnhandledErrorResult
		defer os.RemoveAll(storagePath)
		return testFunc(storagePath)
	}
}

func testWithDatabaseAndStorage(t *testing.T, testFunc func(storage string) error) {
	if err := withDatabase(withStorage(testFunc))(); err != nil {
		t.Error(err.Error())
	}
}
