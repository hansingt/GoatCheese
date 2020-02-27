package web

import (
	"fmt"
	"github.com/hansingt/GoatCheese/datastore"
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
	"regexp"
)

var packageNameRegExp = regexp.MustCompile("[-_.]+")

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}
	return t.templates.ExecuteTemplate(w, name, data)
}

func SetupEchoServer(server *echo.Echo, templatesPath string) error {
	templates := &Template{
		template.Must(template.ParseGlob(fmt.Sprintf("%s/*.html", templatesPath))),
	}
	server.Renderer = templates
	// Root
	server.GET("/", rootView).Name = "root"
	// Repositories
	repos, err := datastore.AllRepositories()
	if err != nil {
		return err
	}
	for _, repo := range repos {
		repoPath := fmt.Sprintf("/%s/", repo.Name())
		projectPath := fmt.Sprintf("%s:project/", repoPath)
		filePath := fmt.Sprintf("%s:file", projectPath)
		server.GET(repoPath, repositoryView(repo)).Name = repo.Name()
		server.POST(repoPath, repositoryPostView(repo)).Name = fmt.Sprintf("%s-post", repo.Name())
		server.GET(projectPath, projectView(repo)).Name = fmt.Sprintf("%s-project", repo.Name())
		server.GET(filePath, projectFileView(repo)).Name = fmt.Sprintf("%s-file", repo.Name())
	}
	return nil
}
