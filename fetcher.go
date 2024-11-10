package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

func FetchServerData(countryID int, cityID int) (string, string, Server) {
	url := "https://api.nordvpn.com/v1/servers/recommendations?filters[servers_technologies][identifier]=wireguard_udp"
	if countryID > 0 {
		url += fmt.Sprintf("&filters[country_id]=%d", countryID)
	}

	// If cityID is specified, we'll fetch all servers and filter for the city
	if cityID <= 0 {
		url += "&limit=1"
	} else {
		url += "&limit=16384"
	}

	resp, err := http.Get(url)
	panicer(err)
	data, err := io.ReadAll(resp.Body)
	panicer(err)
	servers := Servers{}
	err = json.Unmarshal(data, &servers)
	panicer(err)

	if len(servers) == 0 {
		panicer(fmt.Errorf("no servers found for the specified criteria"))
	}

	var selectedServer Server
	if cityID > 0 {
		// Find server in the requested city
		found := false
		for _, server := range servers {
			for _, location := range server.Locations {
				if location.Country.City.ID == cityID {
					selectedServer = server
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			panicer(fmt.Errorf("no servers found in the specified city"))
		}
	} else {
		selectedServer = servers[0]
	}

	hostname := selectedServer.Hostname
	var publicKey string

	for _, technology := range selectedServer.Technologies {
		if technology.Identifier != "wireguard_udp" {
			continue
		}
		publicKey = technology.Metadata[0].Value
	}
	ips, err := net.LookupIP(hostname)
	panicer(err)
	return ips[0].String(), publicKey, selectedServer
}

func getCityID(countryID int, cityName string) int {
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

func fetchOwnPrivateKey(token string) string {
	url := "https://api.nordvpn.com/v1/users/services/credentials"
	req, err := http.NewRequest("GET", url, nil)
	panicer(err)
	req.SetBasicAuth("token", token)
	resp, err := http.DefaultClient.Do(req)
	panicer(err)
	data, err := io.ReadAll(resp.Body)
	panicer(err)
	servers := Creds{}
	err = json.Unmarshal(data, &servers)
	panicer(err)
	return servers.NordlynxPrivateKey
}

type Creds struct {
	ID                 int    `json:"id,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	UpdatedAt          string `json:"updated_at,omitempty"`
	Username           string `json:"username,omitempty"`
	Password           string `json:"password,omitempty"`
	NordlynxPrivateKey string `json:"nordlynx_private_key,omitempty"`
}

type Servers []Server

type Server struct {
	ID           int            `json:"id,omitempty"`
	CreatedAt    string         `json:"created_at,omitempty"`
	UpdatedAt    string         `json:"updated_at,omitempty"`
	Name         string         `json:"name,omitempty"`
	Station      string         `json:"station,omitempty"`
	Ipv6Station  string         `json:"ipv6_station,omitempty"`
	Hostname     string         `json:"hostname,omitempty"`
	Load         int            `json:"load,omitempty"`
	Status       string         `json:"status,omitempty"`
	Locations    []Locations    `json:"locations,omitempty"`
	Technologies []Technologies `json:"technologies,omitempty"`
}
type City struct {
	ID          int     `json:"id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	DNSName     string  `json:"dns_name,omitempty"`
	HubScore    int     `json:"hub_score,omitempty"`
	ServerCount int     `json:"serverCount,omitempty"`
}
type Country struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Code string `json:"code,omitempty"`
	City City   `json:"city,omitempty"`
}
type Locations struct {
	ID        int     `json:"id,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Country   Country `json:"country,omitempty"`
}
type Services struct {
	ID         int    `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}
type Pivot struct {
	TechnologyID int    `json:"technology_id,omitempty"`
	ServerID     int    `json:"server_id,omitempty"`
	Status       string `json:"status,omitempty"`
}
type Technologies struct {
	ID         int        `json:"id,omitempty"`
	Name       string     `json:"name,omitempty"`
	Identifier string     `json:"identifier,omitempty"`
	CreatedAt  string     `json:"created_at,omitempty"`
	UpdatedAt  string     `json:"updated_at,omitempty"`
	Metadata   []Metadata `json:"metadata,omitempty"`
	Pivot      Pivot      `json:"pivot,omitempty"`
}

type Metadata struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}
type Type struct {
	ID         int    `json:"id,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	Title      string `json:"title,omitempty"`
	Identifier string `json:"identifier,omitempty"`
}
type Groups struct {
	ID         int    `json:"id,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	Title      string `json:"title,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	Type       Type   `json:"type,omitempty"`
}
type Values struct {
	ID    int    `json:"id,omitempty"`
	Value string `json:"value,omitempty"`
}
type Specifications struct {
	ID         int      `json:"id,omitempty"`
	Title      string   `json:"title,omitempty"`
	Identifier string   `json:"identifier,omitempty"`
	Values     []Values `json:"values,omitempty"`
}
type IP struct {
	ID      int    `json:"id,omitempty"`
	IP      string `json:"ip,omitempty"`
	Version int    `json:"version,omitempty"`
}
type Ips struct {
	ID        int    `json:"id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	ServerID  int    `json:"server_id,omitempty"`
	IPID      int    `json:"ip_id,omitempty"`
	Type      string `json:"type,omitempty"`
	IP        IP     `json:"ip,omitempty"`
}
