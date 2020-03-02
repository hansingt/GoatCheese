package datastore

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"io"
	"os"
	"path/filepath"
)

/*
IProject defines the interface of a project in the GoatCheese shop.
It defines, that a project needs to have the following properties:

- Name
- ProjectPath
- ProjectFiles

Additionally, it defines, that project files can be added to a project and,
given a filename one can get a specific project file from it.
*/
type IProject interface {
	Name() string                                     // Name returns the name of the project
	ProjectPath() string                              // ProjectPath returns the path on the data storage
	ProjectFiles() ([]IProjectFile, error)            // ProjectFiles returns a slice of all files contained
	GetFile(fileName string) (IProjectFile, error)    // GetFile returns a single file given it's file name
	AddFile(fileName string, content io.Reader) error // AddFile adds a new file to the project
}

type project struct {
	gorm.Model
	db             *datastore `gorm:"-"`
	RepositoryID   uint       `gorm:"unique_index:idx_project;NOT NULL"`
	ProjectName    string     `gorm:"unique_index:idx_project;NOT NULL"`
	RepositoryPath string
}

func newProject(db *datastore, repositoryID uint, projectName string, repositoryPath string) (IProject, error) {
	project := &project{
		db:             db,
		RepositoryID:   repositoryID,
		ProjectName:    projectName,
		RepositoryPath: repositoryPath,
	}
	if _, err := os.Stat(project.ProjectPath()); err != nil {
		if err = os.MkdirAll(project.ProjectPath(), 0750); err != nil {
			return nil, err
		}
	}
	return project, db.Create(project).Error
}

func (p *project) Name() string {
	return p.ProjectName
}

func (p *project) ProjectPath() string {
	return filepath.Join(p.RepositoryPath, p.ProjectName)
}

func (p *project) ProjectFiles() ([]IProjectFile, error) {
	var projectFiles []*projectFile
	err := p.db.Find(&projectFiles, &projectFile{
		ProjectID: p.ID,
	}).Error
	result := make([]IProjectFile, len(projectFiles))
	for i, file := range projectFiles {
		file.db = p.db
		result[i] = file
	}
	return result, err
}

func (p *project) GetFile(fileName string) (IProjectFile, error) {
	file := &projectFile{
		ProjectID: p.ID,
		FileName:  fileName,
	}
	err := p.db.First(file, file).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}
	file.db = p.db
	return file, nil
}

func (p *project) AddFile(fileName string, content io.Reader) error {
	var newFile IProjectFile
	file, err := p.GetFile(fileName)
	if err != nil {
		return err
	}
	if file != nil && file.IsLocked() {
		return fmt.Errorf("file '%s' is currently locked for uploading", fileName)
	} else if file == nil {
		newFile, err = newProjectFile(p.db, p.ID, fileName, p.ProjectPath())
		if err != nil {
			return err
		}
	} else {
		newFile = file
	}
	// Lock the file for uploading
	if err = newFile.Lock(); err != nil {
		if newFile != file {
			_ = newFile.Delete()
		}
		return err
	}
	// Write the contents to the disk
	if err = newFile.Write(content); err != nil {
		if newFile == file {
			// We did modify the file, just unlock it
			_ = file.Unlock()
		} else {
			// We are creating a new file, delete it
			_ = newFile.Delete()
		}
		return err
	}
	// Unlock the file again
	return newFile.Unlock()
}
