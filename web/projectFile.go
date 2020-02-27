package web

import (
	"fmt"
	"github.com/hansingt/GoatCheese/datastore"
	"github.com/labstack/echo/v4"
	"net/http"
)

func projectFileView(repo datastore.IRepository) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		fileName := ctx.Param("file")
		projectName, project, err := getProject(repo, ctx)
		if err != nil {
			return &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  err.Error(),
				Internal: err,
			}
		}
		if project == nil {
			return redirectToPyPi(fmt.Sprintf("%s/%s", projectName, fileName), ctx)
		}
		var file datastore.IProjectFile
		file, err = project.GetFile(fileName)
		if err != nil {
			return &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  err.Error(),
				Internal: err,
			}
		} else if file == nil {
			return &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: fmt.Sprintf("no such file '%s' found in project '%s'", fileName, project.Name()),
			}
		}
		return ctx.File(file.FilePath())
	}
}
