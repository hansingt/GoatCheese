package web

import (
	"fmt"
	"github.com/hansingt/PyPiGo/datastore"
	"github.com/labstack/echo/v4"
	"net/http"
)

func getProject(repo datastore.IRepository, ctx echo.Context) (string, datastore.IProject, error) {
	projectName := ctx.Param("project")
	project, err := repo.GetProject(projectName)
	if err != nil {
		return "", nil, err
	}
	if project == nil {
		projectName = packageNameRegExp.ReplaceAllString(projectName, "-")
		project, err = repo.GetProject(projectName)
		if err != nil {
			return "", nil, err
		}
	}
	return projectName, project, nil
}

func redirectToPyPi(urlPath string, ctx echo.Context) error {
	return ctx.Redirect(http.StatusMovedPermanently, fmt.Sprintf("https://pypi.org/simple/%s/", urlPath))
}

func projectView(repo datastore.IRepository) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		projectName, project, err := getProject(repo, ctx)
		if err != nil {
			return &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  err.Error(),
				Internal: err,
			}
		}
		if project == nil {
			// project not found here. Redirect to PyPi package server
			return redirectToPyPi(projectName, ctx)
		}
		var projectFiles []datastore.IProjectFile
		projectFiles, err = project.ProjectFiles()
		if err != nil {
			return &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  err.Error(),
				Internal: err,
			}
		}
		return ctx.Render(http.StatusOK, "project.html", map[string]interface{}{
			"ProjectName":    project.Name(),
			"ProjectFiles":   projectFiles,
			"RepositoryName": repo.Name(),
		})
	}
}
