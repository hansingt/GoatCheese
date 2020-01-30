package datastore

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/jinzhu/gorm"
	"os"
	"path/filepath"
	"testing"
)

func TestProjectFile_Name(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		fileName := "test.app-15.13.37.42-py2.7.egg"
		check, err := newProjectFile(0, fileName, storage)
		if err != nil {
			return err
		}
		if check.Name() != fileName {
			return errors.New("the file names do not match")
		}
		return nil
	})
}

func TestProjectFile_SetAndGetChecksum(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		fileName := "test.app-15.13.37.42-py2.7.egg"
		check, err := newProjectFile(0, fileName, storage)
		if err != nil {
			return err
		}
		checksum := hex.EncodeToString(sha256.New().Sum(nil))
		if err := check.SetChecksum(checksum); err != nil {
			return err
		}
		if check.Checksum() != checksum {
			return errors.New("the checksums do not match")
		}
		return nil
	})
}

func TestProjectFile_Locking(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		fileName := "test.app-15.13.37.42-py2.7.egg"
		check, err := newProjectFile(0, fileName, storage)
		if err != nil {
			return err
		}
		if check.IsLocked() {
			return errors.New("the file is unexpected locked")
		}
		if err := check.Lock(); err != nil {
			return err
		}
		if !check.IsLocked() {
			return errors.New("the file is not locked")
		}
		if err := check.Unlock(); err != nil {
			return err
		}
		if check.IsLocked() {
			return errors.New("the file is still locked")
		}
		return nil
	})
}

func TestProjectFile_FilePath(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		fileName := "test.app-15.13.37.42-py2.7.egg"
		check, err := newProjectFile(0, fileName, storage)
		if err != nil {
			return err
		}
		if check.FilePath() != filepath.Join(storage, check.Name()) {
			return errors.New("the filepath is not valid")
		}
		return nil
	})
}

func TestProjectFile_Delete(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		fileName := "test.app-15.13.37.42-py2.7.egg"
		file, err := newProjectFile(0, fileName, storage)
		if err != nil {
			return err
		}
		// Check that the file exists
		var check projectFile
		if err := db.Find(&check, file).Error; err != nil {
			return err
		}
		// Now delete it
		if err := file.Delete(); err != nil {
			return err
		}
		// And re-check that it does not exist
		if err := db.Find(&check, file).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		return nil
	})
}

func TestProjectFile_Write(t *testing.T) {
	testWithDatabaseAndStorage(t, func(storage string) error {
		fileName := "test.app-15.13.37.42-py2.7.egg"
		file, err := newProjectFile(0, fileName, storage)
		if err != nil {
			return err
		}
		// Create some random content
		content := make([]byte, 512)
		hashBuilder := sha256.New()
		if _, err := rand.Read(content); err != nil {
			return err
		}
		if _, err := hashBuilder.Write(content); err != nil {
			return err
		}
		contentChecksum := hex.EncodeToString(hashBuilder.Sum(nil))
		// Write them to the file
		if err := file.Write(bytes.NewReader(content)); err != nil {
			return err
		}
		// And then read them to check the content
		outputFile, err := os.Open(file.FilePath())
		if err != nil {
			return err
		}
		//noinspection GoUnhandledErrorResult
		defer outputFile.Close()

		check := make([]byte, len(content))
		if _, err := outputFile.Read(check); err != nil {
			return err
		}
		for i, b := range content {
			if b != check[i] {
				return errors.New("the file contents don't match")
			}
		}
		// Also check, that the checksum has been calculated
		if file.Checksum() != contentChecksum {
			return errors.New("the content checksums don't match")
		}
		return nil
	})
}
