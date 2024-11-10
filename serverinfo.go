package main

import (
	"encoding/json"
	"os"
)

var serverInfoPath = tokenPath + "/current_server.json"

func saveServerInfo(server Server) {
	panicer(os.MkdirAll(tokenPath, 0700))
	data, err := json.MarshalIndent(server, "", "  ")
	panicer(err)
	panicer(os.WriteFile(serverInfoPath, data, 0600))
}

func loadServerInfo() (Server, error) {
	var server Server
	data, err := os.ReadFile(serverInfoPath)
	if err != nil {
		return server, err
	}
	err = json.Unmarshal(data, &server)
	return server, err
}
