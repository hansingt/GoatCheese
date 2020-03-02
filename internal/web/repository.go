package web

import (
	"fmt"
	"github.com/hansingt/GoatCheese/internal/datastore"
	"github.com/labstack/echo/v4"
	"mime/multipart"
	"net/http"
)

func submit(repo datastore.IRepository, form *multipart.Form) (datastore.IProject, error) {
	fieldValues := form.Value["name"]
	if len(fieldValues) != 1 {
		return nil, &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "no (or multiple) field(s) 'name' given in the metadata",
		}
	}
	projectName := packageNameRegExp.ReplaceAllString(fieldValues[0], "-")
	return repo.AddProject(projectName)
}

func fileUpload(repo datastore.IRepository, form *multipart.Form) error {
	prj, err := submit(repo, form)
	if err != nil {
		return err
	}
	files := form.File["content"]
	if len(files) == 0 {
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "no file content uploaded",
		}
	}
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		err = prj.AddFile(fileHeader.Filename, file)
		err2 := file.Close()
		if err != nil {
			return err
		}
		if err2 != nil {
			return err2
		}
	}
	return nil
}

func repositoryView(repo datastore.IRepository) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		projects, err := repo.AllProjects()
		if err != nil {
			return err
		}
		return ctx.Render(http.StatusOK, "repository.html", map[string]interface{}{
			"Repository": repo,
			"Projects":   projects,
		})
	}
}

func repositoryPostView(repo datastore.IRepository) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		form, err := ctx.MultipartForm()
		if err != nil {
			ctx.Error(err)
			return &echo.HTTPError{
				Code:     http.StatusBadRequest,
				Message:  err.Error(),
				Internal: err,
			}
		}
		actions := form.Value[":action"]
		if len(actions) != 1 {
			return &echo.HTTPError{
				Code:    http.StatusBadRequest,
				Message: "no (or multiple) :action found in metadata",
			}
		}
		switch actions[0] {
		case "submit":
			_, err = submit(repo, form)
			return err
		case "file_upload":
			return fileUpload(repo, form)
		default:
			return &echo.HTTPError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("unknown action '%s'", actions[0]),
			}
		}
	}
}
