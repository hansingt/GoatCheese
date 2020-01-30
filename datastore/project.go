package datastore

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"io"
	"os"
	"path/filepath"
)

type IProject interface {
	Name() string
	ProjectPath() string
	ProjectFiles() ([]IProjectFile, error)
	GetFile(fileName string) (IProjectFile, error)
	AddFile(fileName string, content io.Reader) error
}

type project struct {
	gorm.Model
	RepositoryID   uint   `gorm:"unique_index:idx_project;NOT NULL"`
	ProjectName    string `gorm:"unique_index:idx_project;NOT NULL"`
	RepositoryPath string
}

func newProject(repositoryID uint, projectName string, repositoryPath string) (IProject, error) {
	project := &project{
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
	err := db.Find(&projectFiles, &projectFile{
		ProjectID: p.ID,
	}).Error
	result := make([]IProjectFile, len(projectFiles))
	for i, file := range projectFiles {
		result[i] = file
	}
	return result, err
}

func (p *project) GetFile(fileName string) (IProjectFile, error) {
	file := &projectFile{
		ProjectID: p.ID,
		FileName:  fileName,
	}
	err := db.First(file, file).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}
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
		newFile, err = newProjectFile(p.ID, fileName, p.ProjectPath())
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
