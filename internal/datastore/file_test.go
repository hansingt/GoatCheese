package datastore

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
)

type projectFileTestSuite struct {
	TestSuiteWithDatastore
	fileName string
	file     IProjectFile
}

func (suite *projectFileTestSuite) SetupTest() {
	suite.TestSuiteWithDatastore.SetupTest()

	suite.fileName = "test.app-15.13.37.42-py2.7.egg"
	check, err := newProjectFile(suite.db, 0, suite.fileName, suite.storagePath)
	suite.Require().Nil(err, "unable to create a new project file")
	suite.file = check
}

func TestProjectFile(t *testing.T) {
	suite.Run(t, new(projectFileTestSuite))
}

func (suite *projectFileTestSuite) TestName() {
	suite.Require().Equal(suite.file.Name(), suite.fileName, "the file names do not match")
}

func (suite *projectFileTestSuite) TestSetAndGetChecksum() {
	require := suite.Require()
	checksum := hex.EncodeToString(sha256.New().Sum(nil))
	require.Nil(suite.file.SetChecksum(checksum))
	require.Equal(suite.file.Checksum(), checksum, "the checksums do not match")
}

func (suite *projectFileTestSuite) TestLocking() {
	require := suite.Require()
	require.False(suite.file.IsLocked(), "the file is unexpected locked")

	require.Nil(suite.file.Lock(), "could not lock the file")
	require.True(suite.file.IsLocked(), "the file is not locked")

	require.Nil(suite.file.Unlock(), "could not unlock the file")
	require.False(suite.file.IsLocked(), "the file is still locked")
}

func (suite *projectFileTestSuite) TestFilePath() {
	require := suite.Require()
	require.Equal(
		suite.file.FilePath(),
		filepath.Join(suite.storagePath, suite.file.Name()),
		"the filepath is not valid")
}

func (suite *projectFileTestSuite) TestDelete() {
	require := suite.Require()

	// Check that the file exists
	var check projectFile
	require.Nil(suite.db.Find(&check, suite.file).Error, "could not find the file in the database")

	// Now delete it
	require.Nil(suite.file.Delete(), "could not delete the file from the database")

	// And re-check that it does not exist
	require.NotNil(suite.db.Find(&check, suite.file).Error, "Found the file in the database")
}

func (suite *projectFileTestSuite) TestWrite() {
	require := suite.Require()

	// Create some random content
	content := make([]byte, 512)
	hashBuilder := sha256.New()
	_, err := rand.Read(content)
	require.Nil(err, "Error creating some random file content")
	_, err = hashBuilder.Write(content)
	require.Nil(err, "error writing the file contents to the hash builder")
	contentChecksum := hex.EncodeToString(hashBuilder.Sum(nil))

	// Write them to the file
	require.Nil(suite.file.Write(bytes.NewReader(content)), "unable to write the contents to the file")

	// And then read them to check the content
	var outputFile *os.File
	outputFile, err = os.Open(suite.file.FilePath())
	require.Nil(err, "unable to open the target file for reading")
	//noinspection GoUnhandledErrorResult
	defer outputFile.Close()

	check := make([]byte, len(content))
	_, err = outputFile.Read(check)
	require.Nil(err, "unable to read the output file contents")

	for i, b := range content {
		require.Equal(b, check[i], "the file contents don't match")
	}
	// Also check, that the checksum has been calculated
	require.Equal(suite.file.Checksum(), contentChecksum, "the content checksums don't match")
}
