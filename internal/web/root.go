package web

import (
	"github.com/hansingt/GoatCheese/internal/datastore"
	"github.com/labstack/echo/v4"
	"net/http"
)

func rootView(datastore datastore.Datastore) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		repos, err := datastore.AllRepositories()
		if err != nil {
			return &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  err.Error(),
				Internal: err,
			}
		}
		return ctx.Render(http.StatusOK, "repositories.html", map[string]interface{}{
			"Repositories": repos,
		})
	}
}
