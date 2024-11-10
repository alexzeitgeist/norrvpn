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
  up [cc]        Connect to VPN (optionally specify country code)
  down           Disconnect from VPN
  init           Initialize with NordVPN token
  showToken      Display stored token
  listCountries  Show available country codes

Flags:
  --help         Show this help message

Examples:
  norrvpn up gb          Connect to server in Great Britain
  norrvpn listCountries  Show all available country codes
  norrvpn down           Disconnect current session`

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("This program must be run as root (sudo)")
		os.Exit(1)
	}

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
		if flag.NArg() == 2 {
			host, key, server = FetchServerData(getCountryCode(flag.Arg(1)))
		} else {
			host, key, server = FetchServerData(-1)
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
			table.Append([]string{country.Name, country.Code})
		}
		headers := []string{"Country", "Code"}
		table.SetHeader(headers)
		table.Render()
	}
}
