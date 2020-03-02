/*
Package datastore implements the database backend classes required by the GoatCheese application.
*/
package datastore

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Simply import it to be usable as a database backend
	_ "github.com/jinzhu/gorm/dialects/sqlite"   // Simply import it to be usable as a database backend
	"gopkg.in/yaml.v2"
	"os"
)

type IDatastore interface {
	// AllRepositories returns a slice of all repositories defined in the data store.
	AllRepositories() ([]IRepository, error)
	// GetRepository returns the Repository for a given name.
	GetRepository(repositoryName string) (IRepository, error)
	// Close closes the database connection
	Close() error
}

type datastore struct {
	*gorm.DB
}

type indexConfig struct {
	Name  string   `yaml:"name"`
	Bases []string `yaml:"bases"`
}

type databaseConfig struct {
	Driver           string `yaml:"driver"`
	ConnectionString string `yaml:"connection"`
}

type config struct {
	StoragePath string         `yaml:"storagePath"`
	Indexes     []indexConfig  `yaml:"indexes"`
	Database    databaseConfig `yaml:"database"`
}

func readConfigurationFile(configFile string) (*config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	//noinspection GoUnhandledErrorResult
	defer file.Close()

	var cfg config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func setupDatabase(cfg *config) (*datastore, error) {
	db, err := gorm.Open(cfg.Database.Driver, cfg.Database.ConnectionString)
	if err != nil {
		return nil, err
	}
	// Migrate the Schema
	return &datastore{db}, db.AutoMigrate(&projectFile{}).
		AutoMigrate(&project{}).
		AutoMigrate(&repository{}).
		Error
}

func slicesEqual(a, b []string) bool {
	if a == nil && b == nil || a == nil && len(b) == 0 || len(a) == 0 && b == nil {
		return true
	} else if len(a) != len(b) {
		return false
	}
	for _, va := range a {
		found := false
		for _, vb := range b {
			if va == vb {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

/*
New creates a new data store object and set's up the database if required.
It reads the configuration options and repositories to setup from a configuration
file, which's path is given by the `configFile` parameter.

Warning: The data store is a global singleton and thus, this method should only be called once!
*/
func New(configFile string) (IDatastore, error) {
	// Parse the configuration file
	cfg, err := readConfigurationFile(configFile)
	if err != nil {
		return nil, err
	}
	// create the storage path if it does not exist
	_, err = os.Stat(cfg.StoragePath)
	if err != nil {
		err = os.MkdirAll(cfg.StoragePath, 0750)
		if err != nil {
			return nil, err
		}
	}
	// initialize the database connection and tables
	var db *datastore
	db, err = setupDatabase(cfg)
	if err != nil {
		return nil, err
	}
	// add the repositories from the configuration
	err = db.addRepositories(cfg)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func (db *datastore) AllRepositories() ([]IRepository, error) {
	var repositories []*repository
	err := db.Find(&repositories).Error
	if err != nil {
		return nil, err
	}
	result := make([]IRepository, len(repositories))
	for i, repo := range repositories {
		repo.db = db
		result[i] = repo
	}
	return result, nil
}

func (db *datastore) GetRepository(repositoryName string) (IRepository, error) {
	var repo repository
	err := db.Model(&repo).First(&repo, &repository{
		RepositoryName: repositoryName,
	}).Error
	if err != nil {
		return nil, err
	}
	repo.db = db
	return &repo, nil
}

func (db *datastore) Close() error {
	return db.DB.Close()
}

func (db *datastore) addRepositories(cfg *config) error {
	dbRepos, err := db.AllRepositories()
	if err != nil {
		return err
	}
	existingRepos := make(map[string]IRepository, len(dbRepos))
	for _, repo := range dbRepos {
		existingRepos[repo.Name()] = repo
	}
	for _, repo := range cfg.Indexes {
		dbRepo, exists := existingRepos[repo.Name]
		if !exists {
			if _, err := newRepository(db, repo.Name, repo.Bases, cfg.StoragePath); err != nil {
				return err
			}
		} else {
			if cfg.StoragePath != dbRepo.StoragePath() {
				return fmt.Errorf(
					"the storage paths differ: '%s' != '%s'.\n"+
						"If you want to change the storage path, please migrate the pathes in the database",
					cfg.StoragePath, dbRepo.StoragePath())
			}
			bases, err := dbRepo.Bases()
			if err != nil {
				return err
			}
			baseNames := make([]string, len(bases))
			for i, base := range bases {
				baseNames[i] = base.Name()
			}
			if !slicesEqual(baseNames, repo.Bases) {
				var bases []IRepository
				for _, baseName := range repo.Bases {
					base, err := db.GetRepository(baseName)
					if err != nil {
						return err
					}
					bases = append(bases, base)
				}
				if err = dbRepo.SetBases(bases); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
