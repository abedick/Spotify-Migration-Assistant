package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var FILE string
var ALIAS []string
var MAP []string

func Load_config(file string, param_alias []string, param_map []string) map[string]string {

	FILE = file
	ALIAS = param_alias
	MAP = param_map

	settings, err := read_config()

	if err != nil {
		settings = setup()
	}

	return settings
}

func setup() map[string]string {

	tmp_config := make(map[string]string)

	for i, _ := range ALIAS {
		var value string
		fmt.Print(ALIAS[i], ": ")
		fmt.Scanln(&value)
		tmp_config[MAP[i]] = value
	}

	save_config(tmp_config)

	return tmp_config
}

func save_config(conf map[string]string) {

	file, err := os.Create(FILE)

	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for key, value := range conf {
		line := strings.Join([]string{key, value}, "=")
		fmt.Fprintln(writer, line)
	}
	writer.Flush()
}

func read_config() (map[string]string, error) {
	conf := make(map[string]string)

	file, err := os.Open(FILE)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		key_value := strings.Split(scanner.Text(), "=")
		conf[key_value[0]] = key_value[1]
	}
	return conf, nil
}
