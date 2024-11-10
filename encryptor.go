package main

import (
	"encoding/json"
	"os"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func getToken() string {
	data, err := os.ReadFile(tokenFullPath)
	panicer(err)
	var token Token
	err = json.Unmarshal(data, &token)
	panicer(err)
	return token.Token
}

func setToken(token string) {
	panicer(os.MkdirAll(tokenPath, 0700))
	file, err := os.OpenFile(tokenFullPath, os.O_CREATE|os.O_WRONLY, 0600)
	panicer(err)
	tokenObject := Token{Token: token}
	data, err := json.MarshalIndent(tokenObject, "", "  ")
	panicer(err)
	_, err = file.Write(data)
	panicer(err)
}

type Token struct {
	Token string
}
