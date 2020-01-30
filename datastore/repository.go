package datastore

import (
	"github.com/jinzhu/gorm"
	"os"
	"path/filepath"
)

type IRepository interface {
	Name() string
	Bases() ([]IRepository, error)
	AllProjects() ([]IProject, error)
	RepositoryPath() string
	AddProject(projectName string) (IProject, error)
	GetProject(projectName string) (IProject, error)
	StoragePath() string
	SetBases(baseRepositories []IRepository) error
}

type repository struct {
	gorm.Model
	RepositoryName  string        `gorm:"unique_index"`
	RepositoryBases []*repository `gorm:"many2many:repository_bases;association_jointable_foreignkey:parent_id"`
	Storage         string
}

func newRepository(name string, baseNames []string, storagePath string) (IRepository, error) {
	var bases []*repository
	if err := db.Model(&repository{}).Find(&bases, "repository_name IN (?)", baseNames).Error; err != nil {
		return nil, err
	}
	repo := &repository{
		RepositoryName:  name,
		RepositoryBases: bases,
		Storage:         storagePath,
	}
	if _, err := os.Stat(repo.RepositoryPath()); err != nil {
		if err = os.MkdirAll(repo.RepositoryPath(), 0750); err != nil {
			return nil, err
		}
	}
	return repo, db.Model(repo).Create(repo).Error
}

func (r *repository) Name() string {
	return r.RepositoryName
}

func (r *repository) StoragePath() string {
	return r.Storage
}

func (r *repository) RepositoryPath() string {
	return filepath.Join(r.StoragePath(), r.Name())
}

func (r *repository) Bases() ([]IRepository, error) {
	var bases []*repository
	if err := db.Model(r).Association("RepositoryBases").Find(&bases).Error; err != nil {
		return nil, err
	}
	var result []IRepository
	for _, base := range bases {
		result = append(result, base)
	}
	return result, nil
}

func (r *repository) AllProjects() ([]IProject, error) {
	// Find the projects of this repository
	var projects []*project
	if err := db.Find(&projects, &project{RepositoryID: r.ID}).Error; err != nil {
		return nil, err
	}
	projectSet := make(map[string]IProject, len(projects))
	for _, project := range projects {
		projectSet[project.Name()] = project
	}

	// Find the projects of all base repositories
	bases, err := r.Bases()
	if err != nil {
		return nil, err
	}
	for _, base := range bases {
		baseProjects, err := base.AllProjects()
		if err != nil {
			return nil, err
		}
		for _, project := range baseProjects {
			if _, exists := projectSet[project.Name()]; !exists {
				projectSet[project.Name()] = project
			}
		}
	}
	// Then convert the set to a slice
	result := make([]IProject, 0, len(projectSet))
	for _, project := range projectSet {
		result = append(result, project)
	}
	return result, nil
}

func (r *repository) AddProject(projectName string) (IProject, error) {
	if project, err := r.GetProject(projectName); err != nil {
		return nil, err
	} else if project != nil {
		return project, nil
	}
	if project, err := newProject(r.ID, projectName, r.RepositoryPath()); err != nil {
		return nil, err
	} else {
		return project, nil
	}
}

func (r *repository) GetProject(projectName string) (IProject, error) {
	project := &project{
		RepositoryID: r.ID,
		ProjectName:  projectName,
	}
	err := db.Model(project).Find(project, project).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, nil
	}
	return project, nil
}

func (r *repository) SetBases(baseRepositories []IRepository) error {
	var bases []*repository
	for _, base := range baseRepositories {
		bases = append(bases, base.(*repository))
	}
	r.RepositoryBases = bases
	return db.Model(r).Updates(r).Error
}
