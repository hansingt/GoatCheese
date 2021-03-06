package web

import (
	"fmt"
	"github.com/hansingt/GoatCheese/internal/datastore"
	"github.com/labstack/echo/v4"
	"net/http"
)

func projectFileView(repo datastore.Repository) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		fileName := ctx.Param("fileName")
		fileChecksum := ctx.Param("fileChecksum")
		project, err := getProject(repo, ctx)
		if err != nil {
			return err
		}

		var file datastore.ProjectFile
		file, err = project.GetFile(fileName)
		if err != nil {
			return &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  err.Error(),
				Internal: err,
			}
		} else if file == nil || file.Checksum() != fileChecksum {
			return &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: fmt.Sprintf("file not found in project '%s'", project.Name()),
			}
		}
		return ctx.File(file.FilePath())
	}
}
