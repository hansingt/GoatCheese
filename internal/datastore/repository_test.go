package datastore

import (
	"github.com/stretchr/testify/suite"
	"path/filepath"
	"testing"
)

type RepositoryTestSuite struct {
	DatastoreTestSuite
	repoName string
	repo     IRepository
}

func (suite *RepositoryTestSuite) SetupTest() {
	var err error
	suite.DatastoreTestSuite.SetupTest()
	suite.repoName = "apps/15.3"
	suite.repo, err = newRepository(suite.repoName, nil, suite.storagePath)
	suite.Require().Nil(err, "unable to create a new repository")
}

func TestRepository(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (suite *RepositoryTestSuite) TestName() {
	suite.Require().Equal(suite.repoName, suite.repo.Name(), "the repository name is not correct")
}

func (suite *RepositoryTestSuite) TestStoragePath() {
	suite.Require().Equal(suite.storagePath, suite.repo.StoragePath(), "the storage path is not correct")
}

func (suite *RepositoryTestSuite) TestRepositoryPath() {
	suite.Require().Equal(
		filepath.Join(suite.storagePath, suite.repo.Name()),
		suite.repo.RepositoryPath(),
		"the repository path is not correct")
}

func (suite *RepositoryTestSuite) TestSetAndGetBases() {
	var bases []IRepository
	require := suite.Require()

	// Create a base repository
	base, err := newRepository("base", nil, suite.storagePath)
	require.Nil(err, "unable to create a base repository")

	require.Nil(suite.repo.SetBases([]IRepository{base}), "unable to set the repository bases")
	bases, err = suite.repo.Bases()
	require.Nil(err, "unable to get the repository bases")
	require.Equal(1, len(bases), "more than one repository base found")
	require.Equal(base.Name(), bases[0].Name(), "not the correct base has been found")
}

func (suite *RepositoryTestSuite) TestAllProjects() {
	require := suite.Require()

	projects, err := suite.repo.AllProjects()
	require.Nil(err, "unable to get all projects from the repository")
	require.Equal(0, len(projects), "unexpected projects found in repository")

	// Add a project
	projectName := "fuubar"
	_, err = suite.repo.AddProject(projectName)
	require.Nil(err, "unable to add a project to the repository")

	// Now check again
	projects, err = suite.repo.AllProjects()
	require.Nil(err, "unable to get all projects from the repository")
	require.Equal(1, len(projects), "no projects found in repository")
	require.Equal(projectName, projects[0].Name(), "not the correct project found")
}

func (suite *RepositoryTestSuite) TestAllProjectsIncludeBases() {
	projectName := "asdf"
	require := suite.Require()

	// Create a base repository
	base, err := newRepository("base", nil, suite.storagePath)
	require.Nil(err, "unable to create a base repository")
	require.Nil(suite.repo.SetBases([]IRepository{base}))

	_, err = base.AddProject(projectName)
	require.Nil(err, "unable to add the project to the base repository")

	projects, err := suite.repo.AllProjects()
	require.Nil(err, "unable to get all projects from the repository")
	require.Equal(1, len(projects), "no projects found in repository")
	require.Equal(projectName, projects[0].Name(), "not the correct project found")
}

func (suite *RepositoryTestSuite) TestAddAndGetProject() {
	projectName := "fuubar"
	var project IProject
	require := suite.Require()

	// Add a project
	_, err := suite.repo.AddProject(projectName)
	require.Nil(err, "unable to add the project to the repository")

	// And get it again
	project, err = suite.repo.GetProject(projectName)
	require.Nil(err, "unable to get the project from the repository")
	require.NotNil(project, "the project was not found")
	require.Equal(projectName, project.Name(), "not the correct project has been found")
}

func (suite *RepositoryTestSuite) TestAddExistingProject() {
	var projects []IProject
	projectName := "fuubar"
	require := suite.Require()

	// Add a project
	_, err := suite.repo.AddProject(projectName)
	require.Nil(err, "unable to add the project to the repository")

	// Try to add it again
	_, err = suite.repo.AddProject(projectName)
	require.Nil(err, "unable to add the project to the repository")

	// It got added only once
	projects, err = suite.repo.AllProjects()
	require.Nil(err, "unable to get the projects from the repository")
	require.Equal(1, len(projects), "the project got added twice")
}
