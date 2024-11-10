package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func getCountryList() countries {
	resp, err := http.Get("https://api.nordvpn.com/v1/servers/countries")
	panicer(err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	panicer(err)
	c := countries{}
	panicer(json.Unmarshal(data, &c))
	return c
}

func getCountryCode(code string) int {
	for _, country := range getCountryList() {
		if strings.EqualFold(country.Code, code) {
			return country.ID
		}
	}
	return -1
}

func getCityCode(countryID int, cityName string) int {
	for _, country := range getCountryList() {
		if country.ID == countryID {
			for _, city := range country.Cities {
				if strings.EqualFold(city.Name, cityName) {
					return city.ID
				}
			}
		}
	}
	return -1
}

type countries []struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	ServerCount int    `json:"serverCount"`
	Cities      []City `json:"cities"`
}
