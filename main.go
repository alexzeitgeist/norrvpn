// Command wgctrl is a testing utility for interacting with WireGuard via package
// wgctrl.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

const defaultNordvpnAddress = "10.5.0.2/32"

const interfaceName = "norrvpn01"

func main() {
	flag.Parse()
	function := flag.Arg(0)

	switch function {
	case "up":
		var host, key string
		var server Server
		if flag.NArg() == 2 {
			host, key, server = FetchServerData(getCountryCode(flag.Arg(1)))
		} else {
			host, key, server = FetchServerData(-1)
		}
		privateKey := fetchOwnPrivateKey(getToken())
		fmt.Printf("Server name: %s\n", server.Name)
		fmt.Printf("Country: %s (%s)\n", server.Locations[0].Country.Name, server.Locations[0].Country.Code)
		fmt.Printf("City: %s\n", server.Locations[0].Country.City.Name)
		fmt.Printf("Load: %d%%\n", server.Load)
		fmt.Printf("Status: %s\n", server.Status)
		fmt.Printf("Hostname: %s\n", server.Hostname)
		fmt.Printf("WG public key: %s\n", key)
		fmt.Printf("WG private key: %s\n", privateKey)
		execWGup(interfaceName, privateKey, key, host, defaultNordvpnAddress)
	case "down":
		execWGdown(interfaceName, defaultNordvpnAddress)
	case "init":
		token := readSecretInput("Enter TOKEN")
		setToken(token)
	case "showToken":
		fmt.Println(getToken())
	case "listCountries":
		table := tablewriter.NewWriter(os.Stdout)
		for _, country := range getCountryList() {
			table.Append([]string{country.Name, country.Code})
		}
		headers := []string{"Country", "Code"}
		table.SetHeader(headers)
		table.Render()
	}
}
