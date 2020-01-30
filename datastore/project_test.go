package datastore

import (
	"bytes"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestProject_Name(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		projectName := "test-app"
		check, err := newProject(0, projectName, storage)
		if err != nil {
			return err
		}
		if check.Name() != projectName {
			return errors.New("the project name is not correct")
		}
		return nil
	})
}

func TestProject_ProjectPath(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		check, err := newProject(0, "test-app", storage)
		if err != nil {
			return err
		}
		if check.ProjectPath() != filepath.Join(storage, check.Name()) {
			return errors.New("the project path is not valid")
		}
		return nil
	})
}

func TestProject_ProjectFiles(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		check, err := newProject(0, "test-app", storage)
		if err != nil {
			return err
		}
		// Check that no files exist
		if projectFiles, err := check.ProjectFiles(); err != nil {
			return err
		} else if len(projectFiles) > 0 {
			return errors.New("unexpected files found on the project")
		}
		// Add a file
		fileName := "test"
		content := make([]byte, 512)
		if _, err = rand.Read(content); err != nil {
			return err
		}
		if err = check.AddFile(fileName, bytes.NewReader(content)); err != nil {
			return err
		}
		// now check again
		if projectFiles, err := check.ProjectFiles(); err != nil {
			return err
		} else if len(projectFiles) == 0 {
			return errors.New("no files found on the project")
		} else if projectFiles[0].Name() != fileName {
			return errors.New("the file names don't match")
		}
		return nil
	})
}

func TestProject_GetFileNotFound(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		fileName := "test.app-15.13.37.42-py2.7.egg"
		check, err := newProject(0, "test-app", storage)
		if err != nil {
			return err
		}

		var file IProjectFile
		file, err = check.GetFile(fileName)
		if err != nil {
			return err
		} else if file != nil {
			return errors.New("file found in project")
		}
		return nil
	})
}

func TestProject_AddAndGetFile(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		fileName := "test.app-15.13.37.42-py2.7.egg"
		check, err := newProject(0, "test-app", storage)
		if err != nil {
			return err
		}
		content := make([]byte, 512)
		if _, err = rand.Read(content); err != nil {
			return err
		}
		if err = check.AddFile(fileName, bytes.NewReader(content)); err != nil {
			return err
		}

		var file IProjectFile
		if file, err = check.GetFile(fileName); err != nil {
			return err
		} else if file == nil {
			return errors.New("the file has not been found")
		} else if file.IsLocked() {
			return errors.New("the file has not been unlocked")
		}
		return nil
	})
}

func TestProject_AddFileWhileLocked(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		fileName := "test.app-15.13.37.42-py2.7.egg"
		check, err := newProject(0, "test-app", storage)
		if err != nil {
			return err
		}
		content := make([]byte, 512)
		if _, err = rand.Read(content); err != nil {
			return err
		}
		if err = check.AddFile(fileName, bytes.NewReader(content)); err != nil {
			return err
		}

		var file IProjectFile
		file, err = check.GetFile(fileName)
		if err != nil {
			return err
		} else if file == nil {
			return errors.New("the file has not been found")
		}

		if err = file.Lock(); err != nil {
			return err
		}
		if err = check.AddFile(fileName, bytes.NewReader(content)); err == nil {
			return errors.New("no error raised")
		}
		return nil
	})
}

func TestProject_OverwriteFile(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		fileName := "test.app-15.13.37.42-py2.7.egg"
		check, err := newProject(0, "test-app", storage)
		if err != nil {
			return err
		}
		// Add a new file
		content := make([]byte, 512)
		if _, err = rand.Read(content); err != nil {
			return err
		}
		if err = check.AddFile(fileName, bytes.NewReader(content)); err != nil {
			return err
		}

		var file IProjectFile
		if file, err = check.GetFile(fileName); err != nil {
			return err
		} else if file == nil {
			return errors.New("the file has not been found")
		}
		// Now overwrite it with new content
		if _, err = rand.Read(content); err != nil {
			return err
		}
		if err = check.AddFile(fileName, bytes.NewReader(content)); err != nil {
			return err
		}
		// And check that the new file has been stored
		var newFile IProjectFile
		if newFile, err = check.GetFile(fileName); err != nil {
			return err
		} else if newFile == nil {
			return errors.New("the new file has not been found")
		} else if newFile.Checksum() == file.Checksum() {
			return errors.New("the file contents have not been replaced")
		}
		return nil
	})
}

func TestProject_DeleteFileOnError(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		fileName := "test.app-15.13.37.42-py2.7.egg"
		check, err := newProject(0, "test-app", storage)
		if err != nil {
			return err
		}
		// Delete the storage directory to provoke an error
		if err := os.RemoveAll(storage); err != nil {
			return err
		}
		// Try to add a new file
		content := make([]byte, 512)
		if _, err = rand.Read(content); err != nil {
			return err
		}
		if err = check.AddFile(fileName, bytes.NewReader(content)); err == nil {
			return errors.New("no error has been thrown")
		}
		// Check that the file does not exist
		if file, err := check.GetFile(fileName); err != nil {
			return err
		} else if file != nil {
			return errors.New("the file has not been deleted")
		}
		return nil
	})
}

func TestProject_DontDeleteFileOnModifyError(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		fileName := "test.app-15.13.37.42-py2.7.egg"
		check, err := newProject(0, "test-app", storage)
		if err != nil {
			return err
		}
		// Try to add a new file
		content := make([]byte, 512)
		if _, err = rand.Read(content); err != nil {
			return err
		}
		if err = check.AddFile(fileName, bytes.NewReader(content)); err != nil {
			return err
		}
		// Delete the storage directory to provoke an error
		if err := os.RemoveAll(storage); err != nil {
			return err
		}
		// Now try to overwrite it with new content
		if _, err = rand.Read(content); err != nil {
			return err
		}
		if err = check.AddFile(fileName, bytes.NewReader(content)); err == nil {
			return errors.New("no error has been returned")
		}
		// Check that the file still does exist
		if file, err := check.GetFile(fileName); err != nil {
			return err
		} else if file == nil {
			return errors.New("the file has not been found")
		} else if file.IsLocked() {
			return errors.New("the file has not been unlocked")
		}
		return nil
	})
}
