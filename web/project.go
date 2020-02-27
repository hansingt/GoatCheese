package web

import (
	"fmt"
	"github.com/hansingt/GoatCheese/datastore"
	"github.com/labstack/echo/v4"
	"net/http"
)

func getProject(repo datastore.IRepository, ctx echo.Context) (datastore.IProject, error) {
	projectName := ctx.Param("project")
	project, err := repo.GetProject(projectName)
	if err != nil {
		return nil, &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  err.Error(),
			Internal: err,
		}
	}
	if project == nil {
		projectName = packageNameRegExp.ReplaceAllString(projectName, "-")
		project, err = repo.GetProject(projectName)
		if err != nil {
			return nil, &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  err.Error(),
				Internal: err,
			}
		}
	}
	if project == nil {
		return nil, ctx.Redirect(
			http.StatusMovedPermanently,
			fmt.Sprintf("https://pypi.org/simple/%s", projectName))
	}
	return project, nil
}

func projectView(repo datastore.IRepository) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		var projectFiles []datastore.IProjectFile
		project, err := getProject(repo, ctx)
		if err != nil {
			return err
		}

		projectFiles, err = project.ProjectFiles()
		if err != nil {
			return &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  err.Error(),
				Internal: err,
			}
		}
		return ctx.Render(http.StatusOK, "project.html", map[string]interface{}{
			"Repository":   repo,
			"Project":      project,
			"ProjectFiles": projectFiles,
		})
	}
}
