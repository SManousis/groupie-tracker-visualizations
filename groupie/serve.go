package groupie

import (
	"fmt"
	"html/template"
	"net/http"
)

func sub(a, b int) int {
	return a - b
}

func ServeHome(w http.ResponseWriter, r *http.Request, artists *[]Artist) {

	// Prepare the data to be passed to the template
	data := struct {
		Artists []Artist
	}{
		Artists: *artists,
	}
	if r.URL.Path != "/" {
		http.ServeFile(w, r, "templates/error.html")
		return
	}
	// Parse the template file
	tmpl, err := template.New("index.html").Funcs(template.FuncMap{"sub": sub}).ParseFiles("templates/index.html")
	if err != nil {
		// Handle template parsing error
		http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
		return
	}

	// Execute the template with artist data
	if err = tmpl.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}
