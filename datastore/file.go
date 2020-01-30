package datastore

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/jinzhu/gorm"
	"io"
	"os"
	"path/filepath"
)

type IProjectFile interface {
	Name() string
	Checksum() string
	SetChecksum(checksum string) error
	IsLocked() bool
	Lock() error
	Unlock() error
	FilePath() string
	Write(content io.Reader) error
	Delete() error
}

type projectFile struct {
	gorm.Model
	ProjectID    uint   `gorm:"unique_index:idx_project_file;NOT NULL"`
	FileName     string `gorm:"unique_index:idx_project_file;NOT NULL"`
	FileChecksum string
	Locked       bool `gorm:"NOT NULL"`
	ProjectPath  string
}

func newProjectFile(projectID uint, fileName string, projectPath string) (IProjectFile, error) {
	file := &projectFile{
		ProjectID:   projectID,
		FileName:    fileName,
		ProjectPath: projectPath,
		Locked:      false,
	}
	return file, db.Create(file).Error
}

func (f *projectFile) Name() string {
	return f.FileName
}

func (f *projectFile) Checksum() string {
	return f.FileChecksum
}

func (f *projectFile) SetChecksum(checksum string) error {
	f.FileChecksum = checksum
	return db.Model(f).Updates(f).Error
}

func (f *projectFile) IsLocked() bool {
	return f.Locked
}

func (f *projectFile) Lock() error {
	return db.Model(f).Update("Locked", true).Error
}

func (f *projectFile) Unlock() error {
	return db.Model(f).Update("Locked", false).Error
}

func (f *projectFile) FilePath() string {
	return filepath.Join(f.ProjectPath, f.FileName)
}

func (f *projectFile) Write(content io.Reader) error {
	var outputFile *os.File
	var err error
	if outputFile, err = os.OpenFile(f.FilePath(), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0640); err != nil {
		return err
	}
	var n int
	buffer := make([]byte, 100*1024*1024) // 100MiB
	hashBuilder := sha256.New()
	for {
		if n, err = content.Read(buffer); err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		hashBuilder.Write(buffer[:n])
		if _, err = outputFile.Write(buffer[:n]); err != nil {
			return err
		}
	}
	if err = outputFile.Close(); err != nil {
		return err
	}
	return f.SetChecksum(hex.EncodeToString(hashBuilder.Sum(nil)))
}

func (f *projectFile) Delete() error {
	return db.Delete(f).Error
}
