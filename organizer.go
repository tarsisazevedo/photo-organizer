package main

import (
	"encoding/json"
	"fmt"
	"github.com/kellydunn/golang-geo"
	"github.com/rwcarlsen/goexif/exif"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("You should provide a path to organize photos.")
	}
	basePath := os.Args[1]
	photosPaths, err := getPhotos(basePath)
	if err != nil {
		log.Fatal(err)
	}
	if len(photosPaths) == 0 {
		log.Fatal("No photos found")
	}
	organizePhotos(photosPaths)
}

func getPhotos(path string) ([]string, error) {
	var photosPaths []string
	isPhoto, err := filepath.Match("*.jpg", path)
	if err != nil {
		return photosPaths, err
	}
	if isPhoto {
		photosPaths = append(photosPaths, path)
		return photosPaths, nil
	}
	match := fmt.Sprintf("%s/*.jpg", path)
	return filepath.Glob(match)
}

type googleGeocodeResponse struct {
	Results []struct {
		AddressComponents []struct {
			LongName   string   `json:"long_name"`
			Types      []string `json:"types"`
			PostalCode string   `json:"postal_code"`
		} `json:"address_components"`
	}
}

func organizePhotos(paths []string) {
	var wg sync.WaitGroup
	for _, p := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			f, err := os.Open(path)
			if err != nil {
				log.Println(err)
				return
			}
			x, err := exif.Decode(f)
			if err != nil {
				log.Println(err)
				return
			}
			lat, long, _ := x.LatLong()
			p := geo.NewPoint(lat, long)
			geocoder := new(geo.GoogleGeocoder)
			geo.HandleWithSQL()
			geoData, err := geocoder.Request(fmt.Sprintf("latlng=%f,%f", p.Lat(), p.Lng()))
			if err != nil {
				log.Println(err)
			}
			var res googleGeocodeResponse
			if err := json.Unmarshal(geoData, &res); err != nil {
				log.Println(err)
			}
			var city string
			for _, adCompo := range res.Results {
				for _, comp := range adCompo.AddressComponents {
					// See https://developers.google.com/maps/documentation/geocoding/#Types
					// for address types
					for _, compType := range comp.Types {
						if compType == "administrative_area_level_2" {
							city = comp.LongName
							break
						}
					}
				}
			}
			if city != "" {
				err = os.Mkdir(city, 0644)
				if err != nil && !os.IsExist(err) {
					log.Println(err)
					return
				}
				_, filename := filepath.Split(path)
				err = os.Rename(path, fmt.Sprintf("%s/%s", city, filename))
				if err != nil {
					log.Println(err)
					return
				}
			}
		}(p)
	}
	wg.Wait()
}
