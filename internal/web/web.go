package web

import (
	"fmt"
	"github.com/hansingt/GoatCheese/internal/datastore"
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
	"regexp"
)

var packageNameRegExp = regexp.MustCompile("[-_.]+")

type templateRenderer struct {
	templates *template.Template
}

func repositoryUrl(c echo.Context) func(repo datastore.IRepository) string {
	return func(repo datastore.IRepository) string {
		return c.Echo().Reverse(repo.Name())
	}
}

func projectUrl(c echo.Context) func(repo datastore.IRepository, project datastore.IProject) string {
	return func(repository datastore.IRepository, project datastore.IProject) string {
		return c.Echo().Reverse(fmt.Sprintf("%s-project", repository.Name()), project.Name())
	}
}

func projectFileUrl(c echo.Context) func(repo datastore.IRepository, project datastore.IProject, file datastore.IProjectFile) string {
	return func(repo datastore.IRepository, project datastore.IProject, file datastore.IProjectFile) string {
		url := c.Echo().Reverse(
			fmt.Sprintf("%s-file", repo.Name()),
			project.Name(),
			file.Checksum(),
			file.Name())
		return fmt.Sprintf("%s#sha256=%s", url, file.Checksum())
	}
}

/*
Render a template with the given name from the list of templates.
*/
func (t *templateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["repositoryUrl"] = repositoryUrl(c)
		viewContext["projectUrl"] = projectUrl(c)
		viewContext["projectFileUrl"] = projectFileUrl(c)
	}
	return t.templates.ExecuteTemplate(w, name, data)
}

/*
SetupEchoServer sets up the Echo web server to process requests to the GoatCheese shop.
It sets up routes to the endpoints required to be compatible with the python package ecosystem.
*/
func SetupEchoServer(server *echo.Echo, datastore datastore.IDatastore, templatesPath string) error {
	templates := &templateRenderer{
		template.Must(template.ParseGlob(fmt.Sprintf("%s/*.html", templatesPath))),
	}
	server.Renderer = templates

	// Root
	server.GET("/", rootView(datastore)).Name = "root"

	// Repositories
	repos, err := datastore.AllRepositories()
	if err != nil {
		return err
	}
	for _, repo := range repos {
		repoPath := fmt.Sprintf("/%s/", repo.Name())
		projectPath := fmt.Sprintf("%s:project/", repoPath)
		filePath := fmt.Sprintf("%s:fileChecksum/:fileName", projectPath)
		server.GET(repoPath, repositoryView(repo)).Name = repo.Name()
		server.POST(repoPath, repositoryPostView(repo)).Name = fmt.Sprintf("%s-post", repo.Name())
		server.GET(projectPath, projectView(repo)).Name = fmt.Sprintf("%s-project", repo.Name())
		server.GET(filePath, projectFileView(repo)).Name = fmt.Sprintf("%s-file", repo.Name())
	}
	return nil
}
