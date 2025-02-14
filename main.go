package main

import (
	"fmt"
	"groupie/groupie"
	"net/http"
	"os"
)

func main() {
	// Declare a slice to hold the artists data
	var artists []groupie.Artist

	// Fetch the artists once
	if err := groupie.FetchArtists(&artists); err != nil {
		fmt.Println("Error fetching artists:", err)
		return
	}
	//fmt.Println(artists)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serve the homepage on /
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		groupie.ServeHome(w, r, &artists)
	})
	http.HandleFunc("/geolocalization", func(w http.ResponseWriter, r *http.Request) {
		groupie.GeolocalizationHandler(w, r, &artists)
	})
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		groupie.SearchHandler(w, r, &artists)
	})
	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		groupie.ResultHome(w, r, &artists)
	})
	// Get port or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
