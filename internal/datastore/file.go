package datastore

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/jinzhu/gorm"
	"io"
	"log"
	"os"
	"path/filepath"
)

/*
ProjectFile defines the required methods of a project file.
It defines that a project file needs to have the following properties:

- Name
- Checksum
- FilePath
- Lock

Additionally, a file needs to be lockable, can be written and deleted.
*/
type ProjectFile interface {
	Name() string                      // Name returns the name of the project file
	Checksum() string                  // Checksum returns the checksum of the file
	SetChecksum(checksum string) error // SetChecksum sets the checksum of the project file
	IsLocked() bool                    // IsLocked checks whether this file is currently locked by another thread
	Lock() error                       // Lock locks this project file for writing or deletion
	Unlock() error                     // Unlock unlocks this project file for the other threads
	FilePath() string                  // FilePath returns the file path of the project file on the data storage
	Write(content io.Reader) error     // Write writes the contents from the given io.Reader to the file
	Delete() error                     // Delete deletes the project file from the database
}

type projectFile struct {
	gorm.Model
	db           *datastore `gorm:"-"`
	ProjectID    uint       `gorm:"unique_index:idx_project_file;NOT NULL"`
	FileName     string     `gorm:"unique_index:idx_project_file;NOT NULL"`
	FileChecksum string
	Locked       bool `gorm:"NOT NULL"`
	ProjectPath  string
}

func newProjectFile(db *datastore, projectID uint, fileName string, projectPath string) (ProjectFile, error) {
	file := &projectFile{
		db:          db,
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
	return f.db.Model(f).Updates(f).Error
}

func (f *projectFile) IsLocked() bool {
	return f.Locked
}

func (f *projectFile) Lock() error {
	return f.db.Model(f).Update("Locked", true).Error
}

func (f *projectFile) Unlock() error {
	return f.db.Model(f).Update("Locked", false).Error
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
	defer func() {
		if err = outputFile.Close(); err != nil {
			log.Fatalf("Error while closing the output file: %s", err)
		}
	}()

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
		if _, err = hashBuilder.Write(buffer[:n]); err != nil {
			return err
		}
		if _, err = outputFile.Write(buffer[:n]); err != nil {
			return err
		}
	}
	return f.SetChecksum(hex.EncodeToString(hashBuilder.Sum(nil)))
}

func (f *projectFile) Delete() error {
	return f.db.Delete(f).Error
}
