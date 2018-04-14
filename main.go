package main

import (
	"fmt"

	"./config"
)

var CONFIG_FILE = "key.ini"

var global_settings map[string]string

func main() {
	fmt.Println("+---------------------------------------------+")
	fmt.Println("|-- Spotify Migration Assistant --------------|")
	fmt.Println("+---------------------------------------------+")

	config_param_alias := []string{"Client ID", "Client Secret"}
	config_param_mapped := []string{"client_id", "client_secret"}

	global_settings = config.Load_config(CONFIG_FILE, config_param_alias, config_param_mapped)

}
