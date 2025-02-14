package groupie

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func SearchHandler(w http.ResponseWriter, r *http.Request, artists *[]Artist) {
	query := r.URL.Query().Get("q")
	query = strings.ToLower(query)

	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	var listSec []string
	var startsArtist []string
	var startsMembers []string
	var startsLocation []string
	var startsCreation []string
	var startsAlbum []string

	// Check if the search query matches any details
	for _, v := range *artists {
		// Search by artist name
		if strings.HasPrefix(strings.ToLower(v.Name), query) {
			startsArtist = append(startsArtist, v.Name+" - artist/band")
		} else if strings.Contains(strings.ToLower(v.Name), query) {
			listSec = append(listSec, v.Name+" - artist/band")
		}

		// Search by band members
		for _, member := range v.Members {
			if strings.HasPrefix(strings.ToLower(member), query) {
				startsMembers = append(startsMembers, member+" - member of "+v.Name)
			} else if strings.Contains(strings.ToLower(member), query) {
				listSec = append(listSec, member+" - member of "+v.Name)
			}
		}

		// Check for location
		for _, location := range v.Locations {
			if strings.HasPrefix(strings.ToLower(location), query) {
				startsLocation = append(startsLocation, v.Name+" - artist/band performing in: "+location)
			} else if strings.Contains(strings.ToLower(location), query) {
				listSec = append(listSec, v.Name+" - artist/band performing in: "+location)
			}
		}

		// Check if the creation date matches
		if strings.HasPrefix(strconv.Itoa(v.CreationDate), query) {
			startsCreation = append(startsCreation, v.Name+" - artist/band created in"+strconv.Itoa(v.CreationDate))
		} else if strings.Contains(strconv.Itoa(v.CreationDate), query) {
			listSec = append(listSec, v.Name+" - artist/band created in "+strconv.Itoa(v.CreationDate))
		}

		// Check if the first album date matches
		if strings.HasPrefix(strings.ToLower(v.FirstAlbum), query) {
			startsAlbum = append(startsAlbum, v.Name+" - first album released in "+strings.ToLower(v.FirstAlbum))
		} else if strings.Contains(strings.ToLower(v.FirstAlbum), query) {
			listSec = append(listSec, v.Name+" - artist/band first album released in "+strings.ToLower(v.FirstAlbum))
		}
	}

	// Combine the results
	listPrime := append(startsArtist, startsMembers...)
	listPrime = append(listPrime, startsLocation...)
	listPrime = append(listPrime, startsCreation...)
	listPrime = append(listPrime, startsAlbum...)
	list := append(listPrime, listSec...)
	// Return the list of suggestions as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func ResultHome(w http.ResponseWriter, r *http.Request, artists *[]Artist) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Extract the main query
	parts := strings.Split(data.Query, " - ")
	query := parts[0]
	relatedInfo := ""
	if len(parts) > 1 {
		relatedInfo = parts[1] // This will capture "member of <band>" or other info
	}
	queryLower := strings.ToLower(query)

	var matchedArtist *Artist
OuterLoop:
	for _, v := range *artists {
		// Check if the query matches the band or solo artist directly
		if strings.ToLower(v.Name) == queryLower ||
			strings.ToLower(strconv.Itoa(v.CreationDate)) == queryLower ||
			strings.ToLower(v.FirstAlbum) == queryLower {
			matchedArtist = &v
			break OuterLoop
		}

		// If the query matches a band member, ensure we match the correct band
		for _, member := range v.Members {
			if strings.ToLower(member) == queryLower {
				if relatedInfo != "" && strings.Contains(strings.ToLower(relatedInfo), strings.ToLower(v.Name)) {
					matchedArtist = &v
					break OuterLoop
				}
			}
		}

		// Check for exact matches for locations or other criteria
		for _, location := range v.Locations {
			if strings.ToLower(location) == queryLower {
				matchedArtist = &v
				break OuterLoop
			}
		}
	}

	// If no match found, return error
	if matchedArtist == nil {
		http.Error(w, "No artist found", http.StatusNotFound)
		return
	}

	// Return the matched artist's information as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matchedArtist)
}
