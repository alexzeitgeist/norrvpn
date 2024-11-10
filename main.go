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

var helpFlag = flag.Bool("help", false, "Show help information")

const helpText = `Usage: norrvpn [flags] <command> [args]

Commands:
  up [cc] [city] Connect to VPN (optionally specify country code and city)
  down           Disconnect from VPN
  export         Export current connection as WireGuard config
  init           Initialize with NordVPN token
  showToken      Display stored token
  listCountries  Show available country codes

Flags:
  --help         Show this help message

Examples:
  norrvpn up gb london   Connect to server in London, Great Britain
  norrvpn listCountries  Show all available country codes
  norrvpn down           Disconnect current session`

func main() {
	flag.Parse()

	if *helpFlag {
		fmt.Println(helpText)
		os.Exit(0)
	}
	function := flag.Arg(0)

	displayServerInfo := func(server Server) {
		fmt.Printf("Server name: %s\n", server.Name)
		fmt.Printf("Country: %s (%s)\n", server.Locations[0].Country.Name, server.Locations[0].Country.Code)
		fmt.Printf("City: %s\n", server.Locations[0].Country.City.Name)
		fmt.Printf("Load: %d%%\n", server.Load)
		fmt.Printf("Status: %s\n", server.Status)
		fmt.Printf("Hostname: %s\n", server.Hostname)
	}

	if function == "" {
		if !isWGInterfaceExists(interfaceName) {
			if _, err := os.Stat(serverInfoPath); err == nil {
				os.Remove(serverInfoPath)
			}
			fmt.Printf("Not connected\n")
			return
		}

		if server, err := loadServerInfo(); err == nil {
			fmt.Printf("Currently connected to:\n")
			displayServerInfo(server)
		} else {
			fmt.Printf("Connected but server details not available\n")
		}
		return
	}

	switch function {
	case "up":
		if isWGInterfaceExists(interfaceName) {
			fmt.Printf("Error: interface %s already exists. Please disconnect first.\n", interfaceName)
			if server, err := loadServerInfo(); err == nil {
				fmt.Printf("\nCurrently connected to:\n")
				displayServerInfo(server)
			}
			os.Exit(1)
		}

		var host, key string
		var server Server

		if flag.NArg() == 1 {
			host, key, server = FetchServerData(-1, -1)
		} else if flag.NArg() == 2 {
			countryID := getCountryCode(flag.Arg(1))
			if countryID == -1 {
				fmt.Printf("Error: Invalid country code '%s'\n", flag.Arg(1))
				os.Exit(1)
			}
			host, key, server = FetchServerData(countryID, -1)
		} else if flag.NArg() == 3 {
			countryID := getCountryCode(flag.Arg(1))
			if countryID == -1 {
				fmt.Printf("Error: Invalid country code '%s'\n", flag.Arg(1))
				os.Exit(1)
			}
			cityID := getCityCode(countryID, flag.Arg(2))
			if cityID == -1 {
				fmt.Printf("Error: Invalid city name '%s' for country '%s'\n", flag.Arg(2), flag.Arg(1))
				os.Exit(1)
			}
			host, key, server = FetchServerData(countryID, cityID)
		}
		privateKey := fetchOwnPrivateKey(getToken())

		fmt.Printf("Connecting to:\n")
		displayServerInfo(server)
		fmt.Printf("WG public key: %s\n", key)
		fmt.Printf("WG private key: %s\n", privateKey)
		if err := execWGup(interfaceName, privateKey, key, host, defaultNordvpnAddress); err != nil {
			fmt.Printf("Error connecting: %v\n", err)
			os.Exit(1)
		}

		saveServerInfo(server)
	case "down":
		if !isWGInterfaceExists(interfaceName) {
			fmt.Printf("Error: interface %s does not exist\n", interfaceName)
			os.Exit(1)
		}

		if server, err := loadServerInfo(); err == nil {
			fmt.Printf("Disconnecting from %s (%s)...\n",
				server.Locations[0].Country.Name,
				server.Locations[0].Country.Code)
		}
		if err := execWGdown(interfaceName, defaultNordvpnAddress); err != nil {
			fmt.Printf("Error disconnecting: %v\n", err)
			os.Exit(1)
		}
		os.Remove(serverInfoPath)
	case "init":
		token := readSecretInput("Enter TOKEN")
		setToken(token)
	case "showToken":
		fmt.Println(getToken())
	case "listCountries":
		table := tablewriter.NewWriter(os.Stdout)
		for _, country := range getCountryList() {
			for _, city := range country.Cities {
				table.Append([]string{
					country.Name,
					country.Code,
					fmt.Sprintf("%d", country.ID),
					city.Name,
					fmt.Sprintf("%d", city.ID),
				})
			}
		}
		headers := []string{"Country", "Code", "Country ID", "City", "City ID"}
		table.SetHeader(headers)
		table.Render()
	case "export":
		if !isWGInterfaceExists(interfaceName) {
			fmt.Println("Error: Not connected to VPN")
			os.Exit(1)
		}

		server, err := loadServerInfo()
		if err != nil {
			fmt.Printf("Error loading server info: %v\n", err)
			os.Exit(1)
		}

		var publicKey string
		for _, tech := range server.Technologies {
			if tech.Identifier == "wireguard_udp" {
				publicKey = tech.Metadata[0].Value
				break
			}
		}

		privateKey := fetchOwnPrivateKey(getToken())

		fmt.Printf("# wg config for %s (%s)\n", server.Name, server.Locations[0].Country.Code)
		fmt.Printf("[Interface]\n")
		fmt.Printf("Address = %s\n", defaultNordvpnAddress)
		fmt.Printf("PrivateKey = %s\n\n", privateKey)
		fmt.Printf("[Peer]\n")
		fmt.Printf("PublicKey = %s\n", publicKey)
		fmt.Printf("AllowedIPs = 0.0.0.0/0\n")
		fmt.Printf("Endpoint = %s:%s\n", server.Hostname, defaultWGPort)
	}
}
