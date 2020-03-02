package datastore

import (
	"bytes"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

type ProjectTestSuite struct {
	DatastoreTestSuite
	projectName string
	project     IProject
}

func (suite *ProjectTestSuite) SetupTest() {
	var err error
	suite.DatastoreTestSuite.SetupTest()
	suite.projectName = "test-app"
	suite.project, err = newProject(0, suite.projectName, suite.storagePath)
	suite.Require().Nil(err, "unable to create a new project")
}

func TestProject(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}

func (suite *ProjectTestSuite) TestName() {
	suite.Require().Equal(
		suite.projectName,
		suite.project.Name(),
		"the project name is not correct")
}

func (suite *ProjectTestSuite) TestProjectPath() {
	suite.Require().Equal(
		filepath.Join(suite.storagePath, suite.project.Name()),
		suite.project.ProjectPath(),
		"the project path is not valid")
}

func (suite *ProjectTestSuite) TestProjectFiles() {
	require := suite.Require()

	// Check that no files exist
	projectFiles, err := suite.project.ProjectFiles()
	require.Nil(err, "error requesting project files")
	require.Equal(0, len(projectFiles), "unexpected files found on the project")

	// Add a file
	fileName := "test"
	content := make([]byte, 512)
	_, err = rand.Read(content)
	require.Nil(err, "Error creating random file content")

	require.Nil(suite.project.AddFile(fileName, bytes.NewReader(content)), "unable to add the new file")

	// now check again
	projectFiles, err = suite.project.ProjectFiles()
	require.Nil(err, "error requesting project files")
	require.Greater(len(projectFiles), 0, "unexpected files found on the project")
	require.Equal(fileName, projectFiles[0].Name(), "the file names don't match")
}

func (suite *ProjectTestSuite) TestGetFileNotFound() {
	var file IProjectFile
	fileName := "test.app-15.13.37.42-py2.7.egg"
	require := suite.Require()

	file, err := suite.project.GetFile(fileName)
	require.Nil(err, "unable to test the project files")
	require.Nil(file, "file found in project")
}

func (suite *ProjectTestSuite) TestAddAndGetFile() {
	var file IProjectFile
	fileName := "test.app-15.13.37.42-py2.7.egg"
	require := suite.Require()

	content := make([]byte, 512)
	_, err := rand.Read(content)
	require.Nil(err, "Error creating random file content")
	require.Nil(suite.project.AddFile(fileName, bytes.NewReader(content)), "error adding the project file")

	file, err = suite.project.GetFile(fileName)
	require.Nil(err, "error getting the project file")
	require.NotNil(file, "the file has not been found")
	require.False(file.IsLocked(), "the file has not been unlocked")
}

func (suite *ProjectTestSuite) TestAddFileWhileLocked() {
	var file IProjectFile
	fileName := "test.app-15.13.37.42-py2.7.egg"
	require := suite.Require()

	content := make([]byte, 512)
	_, err := rand.Read(content)
	require.Nil(err, "Error creating random file content")
	require.Nil(suite.project.AddFile(fileName, bytes.NewReader(content)), "error adding the project file")

	file, err = suite.project.GetFile(fileName)
	require.Nil(err, "error getting the project file")
	require.NotNil(file, "the file has not been found")

	require.Nil(file.Lock(), "could not lock the file")
	require.NotNil(suite.project.AddFile(fileName, bytes.NewReader(content)), "no error raised")
}

func (suite *ProjectTestSuite) TestOverwriteFile() {
	fileName := "test.app-15.13.37.42-py2.7.egg"
	var file IProjectFile
	var newFile IProjectFile
	require := suite.Require()

	// Add a new file
	content := make([]byte, 512)
	_, err := rand.Read(content)
	require.Nil(err, "Error creating random file content")
	require.Nil(suite.project.AddFile(fileName, bytes.NewReader(content)), "error adding the project file")
	file, err = suite.project.GetFile(fileName)
	require.Nil(err, "error getting the project file")
	require.NotNil(file, "the file has not been found")

	// Now overwrite it with new content
	_, err = rand.Read(content)
	require.Nil(err, "Error creating random file content")
	require.Nil(suite.project.AddFile(fileName, bytes.NewReader(content)), "error adding the project file")

	// And check that the new file has been stored
	newFile, err = suite.project.GetFile(fileName)
	require.Nil(err, "error getting the project file")
	require.NotNil(newFile, "the file has not been found")
	require.NotEqual(newFile.Checksum(), file.Checksum(), "the file contents have not been replaced")
}

func (suite *ProjectTestSuite) TestDeleteFileOnError() {
	var file IProjectFile
	fileName := "test.app-15.13.37.42-py2.7.egg"
	require := suite.Require()

	// Delete the storage directory to provoke an error
	require.Nil(os.RemoveAll(suite.storagePath), "could not delete the storage directory")

	// Try to add a new file
	content := make([]byte, 512)
	_, err := rand.Read(content)
	require.Nil(err, "Error creating random file content")
	require.NotNil(suite.project.AddFile(fileName, bytes.NewReader(content)), "no error has been raised")
	file, err = suite.project.GetFile(fileName)
	require.Nil(err, "error getting the project file")
	require.Nil(file, "the file has not been deleted")
}

func (suite *ProjectTestSuite) TestDontDeleteFileOnModifyError() {
	var file IProjectFile
	fileName := "test.app-15.13.37.42-py2.7.egg"
	require := suite.Require()

	// Try to add a new file
	content := make([]byte, 512)
	_, err := rand.Read(content)
	require.Nil(err, "Error creating random file content")
	require.Nil(suite.project.AddFile(fileName, bytes.NewReader(content)), "error adding the project file")

	// Delete the storage directory to provoke an error
	require.Nil(os.RemoveAll(suite.storagePath), "could not delete the storage directory")

	// Try to add a new file
	_, err = rand.Read(content)
	require.Nil(err, "Error creating random file content")
	require.NotNil(suite.project.AddFile(fileName, bytes.NewReader(content)), "no error has been raised")
	file, err = suite.project.GetFile(fileName)
	require.Nil(err, "error getting the project file")
	require.NotNil(file, "the file has not been found")
	require.False(file.IsLocked(), "the file has not been unlocked")
}
