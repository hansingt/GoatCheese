package datastore

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestRepository_Name(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		repoName := "apps/15.3"
		if check, err := newRepository(repoName, nil, storage); err != nil {
			return err
		} else if check.Name() != repoName {
			return errors.New("the repository name is not correct")
		}
		return nil
	})
}

func TestRepository_StoragePath(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		repoName := "apps/15.3"
		if check, err := newRepository(repoName, nil, storage); err != nil {
			return err
		} else if check.StoragePath() != storage {
			return errors.New("the storage path is not correct")
		}
		return nil
	})
}

func TestRepository_RepositoryPath(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		repoName := "apps/15.3"
		if check, err := newRepository(repoName, nil, storage); err != nil {
			return err
		} else if check.RepositoryPath() != filepath.Join(storage, check.Name()) {
			return errors.New("the repository path is not correct")
		}
		return nil
	})
}

func TestRepository_SetAndGetBases(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		// Create a base repository
		base, err := newRepository("base", nil, storage)
		if err != nil {
			return err
		}
		repo, err := newRepository("test", nil, storage)
		if err != nil {
			return err
		}
		if err := repo.SetBases([]IRepository{base}); err != nil {
			return err
		}
		if bases, err := repo.Bases(); err != nil {
			return err
		} else if len(bases) != 1 {
			return errors.New("not all (or too many) repository bases found")
		} else if bases[0].Name() != base.Name() {
			return errors.New("not the correct base has been found")
		}
		return nil
	})
}

func TestRepository_AllProjects(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		check, err := newRepository("test", nil, storage)
		if err != nil {
			return err
		}
		var projects []IProject
		if projects, err = check.AllProjects(); err != nil {
			return err
		} else if len(projects) != 0 {
			return errors.New("unexpected projects found in repository")
		}
		// Add a project
		projectName := "fuubar"
		if _, err = check.AddProject(projectName); err != nil {
			return err
		}
		// Now check again
		if projects, err = check.AllProjects(); err != nil {
			return err
		} else if len(projects) != 1 {
			return errors.New("no projects found in repository")
		}
		return nil
	})
}

func TestRepository_AllProjectsIncludeBases(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		base, err := newRepository("base", nil, storage)
		if err != nil {
			return err
		}
		projectName := "asdf"
		if _, err = base.AddProject(projectName); err != nil {
			return err
		}
		check, err := newRepository("test", []string{base.Name()}, storage)
		if err != nil {
			return err
		}
		if projects, err := check.AllProjects(); err != nil {
			return err
		} else if len(projects) != 1 {
			return errors.New("no projects found in the repository")
		} else if projects[0].Name() != projectName {
			return errors.New("not the correct project has been found")
		}
		return nil
	})
}

func TestRepository_AddAndGetProject(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		check, err := newRepository("test", nil, storage)
		if err != nil {
			return err
		}
		// Add a project
		projectName := "fuubar"
		if _, err = check.AddProject(projectName); err != nil {
			return err
		}
		// And get it again
		if project, err := check.GetProject(projectName); err != nil {
			return err
		} else if project == nil {
			return errors.New("the project has not been found")
		} else if project.Name() != projectName {
			return errors.New("not the correct project found")
		}
		return nil
	})
}

func TestRepository_AddExistingProject(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		repo, err := newRepository("test", nil, storage)
		if err != nil {
			return err
		}
		// Add a project
		projectName := "fuubar"
		if _, err := repo.AddProject(projectName); err != nil {
			return err
		}
		// Try to add it again
		if _, err := repo.AddProject(projectName); err != nil {
			return err
		}
		// It got added only once
		if projects, err := repo.AllProjects(); err != nil {
			return nil
		} else if len(projects) > 1 {
			return errors.New("the project got added twice")
		}
		return nil
	})
}
