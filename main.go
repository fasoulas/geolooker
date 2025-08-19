package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// ----------- Response structs -----------

type GoogleGeocodeResponse struct {
	Results []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
	Status string `json:"status"`
}

type OSMGeocodeResponse []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type PositionstackResponse struct {
	Data []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"data"`
}

type OpenCageResponse struct {
	Results []struct {
		Geometry struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geometry"`
	} `json:"results"`
}

type LocationIQResponse []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type MapQuestResponse struct {
	Results []struct {
		Locations []struct {
			LatLng struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"latLng"`
		} `json:"locations"`
	} `json:"results"`
	Info struct {
		Statuscode int `json:"statuscode"`
	} `json:"info"`
}

// ----------- Output struct -----------

type GeocodeResult struct {
	Provider  string  `json:"provider"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// ----------- Helper functions -----------

func printJSON(provider, address string, lat, lng float64) {
	res := GeocodeResult{
		Provider:  provider,
		Address:   address,
		Latitude:  lat,
		Longitude: lng,
	}
	data, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(data))
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// ----------- Provider functions -----------

func geocodeGoogle(address string) (float64, float64, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return 0, 0, fmt.Errorf("GOOGLE_API_KEY not set")
	}
	endpoint := "https://maps.googleapis.com/maps/api/geocode/json"
	resp, err := http.Get(fmt.Sprintf("%s?address=%s&key=%s", endpoint, url.QueryEscape(address), apiKey))
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var result GoogleGeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, err
	}
	if result.Status != "OK" || len(result.Results) == 0 {
		return 0, 0, fmt.Errorf("no results (status: %s)", result.Status)
	}
	lat := result.Results[0].Geometry.Location.Lat
	lng := result.Results[0].Geometry.Location.Lng
	return lat, lng, nil
}

func geocodeOSM(address string) (float64, float64, error) {
	endpoint := "https://nominatim.openstreetmap.org/search"
	query := fmt.Sprintf("%s?q=%s&format=json&limit=1", endpoint, url.QueryEscape(address))
	req, _ := http.NewRequest("GET", query, nil)
	req.Header.Set("User-Agent", "Go-Geocoder/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var result OSMGeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, err
	}
	if len(result) == 0 {
		return 0, 0, fmt.Errorf("no results")
	}
	lat, lng := parseFloat(result[0].Lat), parseFloat(result[0].Lon)
	return lat, lng, nil
}

func geocodePositionstack(address string) (float64, float64, error) {
	apiKey := os.Getenv("POSITIONSTACK_KEY")
	if apiKey == "" {
		return 0, 0, fmt.Errorf("POSITIONSTACK_KEY not set")
	}
	endpoint := "http://api.positionstack.com/v1/forward"
	query := fmt.Sprintf("%s?access_key=%s&query=%s&limit=1", endpoint, apiKey, url.QueryEscape(address))
	resp, err := http.Get(query)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var result PositionstackResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, err
	}
	if len(result.Data) == 0 {
		return 0, 0, fmt.Errorf("no results")
	}
	return result.Data[0].Latitude, result.Data[0].Longitude, nil
}

func geocodeOpenCage(address string) (float64, float64, error) {
	apiKey := os.Getenv("OPENCAGE_KEY")
	if apiKey == "" {
		return 0, 0, fmt.Errorf("OPENCAGE_KEY not set")
	}
	endpoint := "https://api.opencagedata.com/geocode/v1/json"
	query := fmt.Sprintf("%s?q=%s&key=%s&limit=1", endpoint, url.QueryEscape(address), apiKey)
	resp, err := http.Get(query)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var result OpenCageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, err
	}
	if len(result.Results) == 0 {
		return 0, 0, fmt.Errorf("no results")
	}
	lat := result.Results[0].Geometry.Lat
	lng := result.Results[0].Geometry.Lng
	return lat, lng, nil
}

func geocodeLocationIQ(address string) (float64, float64, error) {
	apiKey := os.Getenv("LOCATIONIQ_KEY")
	if apiKey == "" {
		return 0, 0, fmt.Errorf("LOCATIONIQ_KEY not set")
	}
	endpoint := "https://us1.locationiq.com/v1/search.php"
	query := fmt.Sprintf("%s?key=%s&q=%s&format=json&limit=1", endpoint, apiKey, url.QueryEscape(address))
	resp, err := http.Get(query)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var result LocationIQResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, err
	}
	if len(result) == 0 {
		return 0, 0, fmt.Errorf("no results")
	}
	lat, lng := parseFloat(result[0].Lat), parseFloat(result[0].Lon)
	return lat, lng, nil
}

func geocodeMapQuest(address string) (float64, float64, error) {
	apiKey := os.Getenv("MAPQUEST_KEY")
	if apiKey == "" {
		return 0, 0, fmt.Errorf("MAPQUEST_KEY not set")
	}
	endpoint := "http://www.mapquestapi.com/geocoding/v1/address"
	query := fmt.Sprintf("%s?key=%s&location=%s", endpoint, apiKey, url.QueryEscape(address))
	resp, err := http.Get(query)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var result MapQuestResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, err
	}
	if result.Info.Statuscode != 0 || len(result.Results) == 0 || len(result.Results[0].Locations) == 0 {
		return 0, 0, fmt.Errorf("no results")
	}
	lat := result.Results[0].Locations[0].LatLng.Lat
	lng := result.Results[0].Locations[0].LatLng.Lng
	return lat, lng, nil
}

// ----------- Main function -----------

func main() {
	provider := flag.String("provider", "osm", "Primary geocoding provider")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: geocode --provider <provider> <address>")
		os.Exit(1)
	}

	address := strings.Join(flag.Args(), " ")

	// List of providers
	providers := []struct {
		name  string
		fn    func(string) (float64, float64, error)
		isAPI bool
		env   string
	}{
		{"google", geocodeGoogle, true, "GOOGLE_API_KEY"},
		{"positionstack", geocodePositionstack, true, "POSITIONSTACK_KEY"},
		{"opencage", geocodeOpenCage, true, "OPENCAGE_KEY"},
		{"locationiq", geocodeLocationIQ, true, "LOCATIONIQ_KEY"},
		{"mapquest", geocodeMapQuest, true, "MAPQUEST_KEY"},
		{"osm", geocodeOSM, false, ""},
	}

	// Find selected provider
	var selected *struct {
		name  string
		fn    func(string) (float64, float64, error)
		isAPI bool
		env   string
	}
	for _, p := range providers {
		if p.name == *provider {
			selected = &p
			break
		}
	}

	// Warnings for invalid provider or missing API key
	if selected == nil {
		fmt.Fprintf(os.Stderr, "Warning: provider '%s' not recognized. Falling back to available providers.\n", *provider)
	} else if selected.isAPI && os.Getenv(selected.env) == "" {
		fmt.Fprintf(os.Stderr, "Warning: API key for provider '%s' not set in environment variable %s. Falling back to other providers.\n", selected.name, selected.env)
	}

	// Reorder: selected first (if valid), then the rest
	var ordered []struct {
		name  string
		fn    func(string) (float64, float64, error)
		isAPI bool
		env   string
	}
	if selected != nil {
		ordered = append(ordered, *selected)
	}
	for _, p := range providers {
		if selected == nil || p.name != selected.name {
			ordered = append(ordered, p)
		}
	}

	// Try providers until one succeeds
	for _, p := range ordered {
		lat, lng, err := p.fn(address)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Provider %s failed: %v\n", p.name, err)
			continue
		}
		printJSON(p.name, address, lat, lng)
		return
	}

	fmt.Fprintln(os.Stderr, "All providers failed")
	os.Exit(1)
}
