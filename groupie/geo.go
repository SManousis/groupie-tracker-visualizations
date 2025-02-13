package groupie

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func findMatchedArtist(query string, artists []Artist) (*Artist, error) {
	queryLower := strings.ToLower(query)

	for _, v := range artists {
		if strings.ToLower(v.Name) == queryLower ||
			strings.ToLower(strconv.Itoa(v.CreationDate)) == queryLower ||
			strings.ToLower(v.FirstAlbum) == queryLower {
			return &v, nil
		}

		for _, member := range v.Members {
			if strings.ToLower(member) == queryLower {
				return &v, nil
			}
		}

		for _, location := range v.Locations {
			if strings.ToLower(location) == queryLower {
				return &v, nil
			}
		}
	}

	return nil, nil
}

func GeolocalizationHandler(w http.ResponseWriter, r *http.Request, artists *[]Artist) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}
	// fmt.Println("Geolocation request received for:", query)
	// Find the matched artist
	matchedArtist, err := findMatchedArtist(query, *artists)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error finding artist: %v", err), http.StatusInternalServerError)
		return
	}

	if matchedArtist == nil {
		http.Error(w, "No artist found", http.StatusNotFound)
		return
	}
	// fmt.Println("Matched artist:", matchedArtist.Name)
	// fmt.Println("Artist locations:", matchedArtist.Locations)
	// Process the artist's locations
	// 1. Pair Dates with Locations
	type LocationDate struct {
		Location string
		Date     time.Time
	}

	var locationDates []LocationDate
	layout := "02-01-2006" // Assuming the dates are in "DD-MM-YYYY" format

	for i, location := range matchedArtist.Locations {
		if i < len(matchedArtist.Dates) {
			// Clean the date string by trimming unwanted characters (like '*')
			cleanedDate := strings.TrimSpace(matchedArtist.Dates[i])
			cleanedDate = strings.TrimPrefix(cleanedDate, "*") // Remove '*' if present

			parsedDate, err := time.Parse(layout, cleanedDate)
			if err != nil {
				fmt.Println("Error parsing date:", cleanedDate, err)
				continue
			}

			locationDates = append(locationDates, LocationDate{
				Location: location,
				Date:     parsedDate,
			})
		}
	}

	// 2. Sort Locations by Parsed Dates
	sort.Slice(locationDates, func(i, j int) bool {
		return locationDates[i].Date.Before(locationDates[j].Date)
	})

	// 3. Process Sorted Locations
	uniqueLocations := make(map[string]bool)
	var sortedLocations []string

	for _, ld := range locationDates {
		editedLocation := strings.ReplaceAll(ld.Location, "_", " ")
		formattedLocation := strings.ReplaceAll(editedLocation, "-", ", ")

		if !uniqueLocations[formattedLocation] {
			uniqueLocations[formattedLocation] = true
			sortedLocations = append(sortedLocations, formattedLocation)
		}
	}

	type GeocodedLocation struct {
		Location string  `json:"location"`
		Lat      float64 `json:"lat"`
		Lon      float64 `json:"lon"`
	}

	var results []GeocodedLocation

	for i := range sortedLocations {
		geoURL := fmt.Sprintf(
			"https://api.geoapify.com/v1/geocode/search?text=%s&apiKey=c62505595d2745efb44a0e32dc6e86bb",
			url.QueryEscape(sortedLocations[i]),
		)
		// fmt.Println("Geoapify Request URL:", geoURL)
		client := &http.Client{}
		req, err := http.NewRequest("GET", geoURL, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			fmt.Println("Error making request:", err)
			continue
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			continue
		}

		type GeoapifyResponse struct {
			Features []struct {
				Geometry struct {
					Coordinates []float64 `json:"coordinates"`
				} `json:"geometry"`
			} `json:"features"`
		}

		var geoResponse GeoapifyResponse
		if err := json.Unmarshal(body, &geoResponse); err != nil {
			fmt.Println("Error parsing Geoapify response:", err)
			continue
		}

		if len(geoResponse.Features) > 0 {
			lat := geoResponse.Features[0].Geometry.Coordinates[1]
			lon := geoResponse.Features[0].Geometry.Coordinates[0]
			results = append(results, GeocodedLocation{
				Location: sortedLocations[i],
				Lat:      lat,
				Lon:      lon,
			})
		} else {
			fmt.Printf("No coordinates found for location: %s\n", sortedLocations[i])
		}
	}

	// Send results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
