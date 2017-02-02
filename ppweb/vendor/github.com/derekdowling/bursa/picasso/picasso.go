// This package handles the building of layouts, partials, and templates
// into a renderable package which can then be written to an HTTP response.
package picasso

import (
	"html/template"
	"net/http"
	"path"
	"path/filepath"

	"github.com/derekdowling/bursa/config"
	"runtime"
)

// Call this to render a template via a Response Writer.
// This automatically sets headers to return http.StatusOK
func Render(w http.ResponseWriter, layout string, view string, pipeline interface{}) {
	RenderWithCode(w, layout, view, pipeline, http.StatusOK)
}

func RenderWithCode(w http.ResponseWriter, layout string, view string, pipeline interface{}, status int) {
	
	template := buildTemplate(layout, view)

	// if the user has provided a non-success code, manually fire the header
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	// Provides some visibility into template execution errors.
	if err := template.Execute(w, pipeline); err != nil {
		//TODO: render a 500 with Picasso
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
// Handles the magic of creating a template
func buildTemplate(layout string, view string) *template.Template {

	// load our base views folder
	template_dir := getTemplateRoot()

	// apply expected extensions to the base layout and view
	layout_path := path.Join(template_dir, layout+".tmpl")
	view_path := path.Join(template_dir, view+".tmpl")

	// collect any partial templates that can be plugged in
	partials := findPartials(layout_path)

	// combine our layout, view, and plugin any partials and return the built template
	return combineTemplates(layout_path, view_path, partials)
}

// Plugs in all of our templates to the base layout
func combineTemplates(layout string, view string, partials []string) *template.Template {
	templates := []string{layout, view}
	templates = append(templates, partials...)
	return template.Must(template.ParseFiles(templates...))
}

// Creates a relative path to our templates folder
func getTemplateRoot() string {
	_, filename, _, _ := runtime.Caller(1)
	base_path := path.Join(path.Dir(filename), "/../")
	return path.Join(base_path, config.App.GetString("paths.templates"))
}

// Searches the folder that the layout is defined in for a "/partials" folder
// and parses partial file names into a slice if the folder exists and it is
// populated
func findPartials(layout_path string) []string {

	expected_partial_dir := path.Join(path.Dir(layout_path), "partials")

	// now do a dir listing
	files, err := filepath.Glob(expected_partial_dir + "/*")
	if err != nil {
		files = make([]string, 2)
	}

	return files
}
