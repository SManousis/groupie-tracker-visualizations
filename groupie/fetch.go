package groupie

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Artist struct {
	ID           int               `json:"id"`
	Image        string            `json:"image"`
	Name         string            `json:"name"`
	Members      []string          `json:"members"`
	CreationDate int               `json:"creationDate"`
	FirstAlbum   string            `json:"firstAlbum"`
	Locations    []string          `json:"-"`
	Dates        []string          `json:"-"`
	Relation     map[string]string `json:"-"`
}

const baseUrl = "https://groupietrackers.herokuapp.com/api/"

var endpoints = []string{"locations/", "dates/", "relation/"}

func FetchArtists(artists *[]Artist) error {
	response, err := http.Get(baseUrl + "artists")
	if err != nil {
		return err
	}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(artists)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for counter := range *artists {
		for _, endpoint := range endpoints {
			wg.Add(1)

			go func(artist *Artist, endpoint string) {
				defer wg.Done()
				if err := UpdateArtistStruct(artist, endpoint); err != nil {
					fmt.Println("Error updating artist:", err)
				}
			}(&(*artists)[counter], endpoint)

		}
	}
	wg.Wait()
	return nil
}

func UpdateArtistStruct(artist *Artist, endpoint string) error {
	response, err := http.Get(baseUrl + endpoint + strconv.Itoa(artist.ID))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var rawData map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&rawData); err != nil {
		return err
	}

	switch endpoint {
	case "locations/":
		if locations, exists := rawData["locations"].([]interface{}); exists {
			artist.Locations = make([]string, len(locations))
			for i, v := range locations {
				artist.Locations[i] = strings.ToTitle(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprint(v), "_", " "), "-", ", "))

			}
		}
	case "dates/":
		if dates, exists := rawData["dates"].([]interface{}); exists {
			artist.Dates = make([]string, len(dates))
			for i, v := range dates {
				date := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprint(v), "[", ""), "]", ""), " ", ", ")
				artist.Dates[i] = date
			}
		}
	case "relation/":
		if datesLocations, exists := rawData["datesLocations"].(map[string]interface{}); exists {
			artist.Relation = make(map[string]string)
			for k, v := range datesLocations {
				location := strings.ReplaceAll(strings.ReplaceAll(fmt.Sprint(k), "_", " "), "-", ", ")
				date := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprint(v), "[", ""), "]", ""), " ", ", ")
				artist.Relation[strings.ToTitle(location)] = date
			}
		}
	}

	return nil
}
