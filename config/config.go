/*
 *
 * Package Config
 *
 * Desc		: An easy configuration system that stores all neccessary global
 *			  program variables in a hash table that can quickly be updated,
 *			  read, and saved to file.
 *
 *			  Configuration files are not meant to be shared and are in no way
 *			  encrypted.
 *
 * Author	: Abe Dick <abedick8213@gmail.com>
 * Date		: April, May 2018
 */

package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

/*
 * Load Configuration
 *
 * * PUBLIC METHOD * *
 *
 * Desc : Attempts to load into a hash table configuration settings read in
 *		  from a settings file with the name 'file'. If this can be done, the
 *		  data is returned, else a setup procedure is ran to collect neccessary
 *		  configuration details from the user.
 *
 */
func Load_config(file string, param_alias []string, param_map []string) map[string]string {

	settings, err := read_config(file, param_alias, param_map)

	if err != nil {
		settings = setup(file, param_alias, param_map)
	}

	return settings
}

/*
 * Update Configuration
 *
 * * PUBLIC METHOD * *
 *
 * Desc : Adds a variable and value to the configuration and saves the updates
 * 		  to file.
 *
 */
func Update_config(file string, config map[string]string, param_name string, param_value string) map[string]string {

	config[param_name] = param_value
	save_config(file, config)

	return config
}

/*
 * Setup
 *
 * * PRIVATE METHOD * *
 *
 * Desc : Gathers all parameters specified in param map from the user.
 *
 */
func setup(file string, param_alias []string, param_map []string) map[string]string {

	tmp_config := make(map[string]string)

	for i, _ := range param_alias {
		var value string
		fmt.Print(param_alias[i], ": ")
		fmt.Scanln(&value)
		tmp_config[param_map[i]] = value
	}

	save_config(file, tmp_config)

	return tmp_config
}

/*
 * Save Configuration
 *
 * * PRIVATE METHOD * *
 *
 * Desc : Saves all parameters stored in the hash table in plain text.
 *
 * Future update may contain encryption for saved parameters.
 *
 */
func save_config(file string, conf map[string]string) {

	write_file, err := os.Create(file)

	if err != nil {
		panic(err)
	}
	defer write_file.Close()

	writer := bufio.NewWriter(write_file)

	for key, value := range conf {
		line := strings.Join([]string{key, value}, "=")
		fmt.Fprintln(writer, line)
	}
	writer.Flush()
}

/*
 * Read Configuration
 *
 * * PRIVATE METHOD * *
 *
 * Desc : Attempts to load all configuration parameters from the specified save
 *		  file and into a hash table. If a parameter is missing from the read
 *		  in file but is present in the list of expected parameters, it is now
 *		  gathered from the user. If a parameter is read and not an expected
 *		  parameter it is ignored.
 *
 */
func read_config(file string, param_alias []string, param_map []string) (map[string]string, error) {
	conf := make(map[string]string)

	read_file, err := os.Open(file)

	if err != nil {
		return nil, err
	}
	defer read_file.Close()

	scanner := bufio.NewScanner(read_file)

	for scanner.Scan() {
		key_value := strings.Split(scanner.Text(), "=")
		conf[key_value[0]] = key_value[1]
	}

	update := false

	/* Make sure that all config variables needed were read */
	for i, value := range param_map {

		if _, ok := conf[value]; !ok {

			var value string
			fmt.Print(param_alias[i], ": ")
			fmt.Scanln(&value)
			conf[param_map[i]] = value

			update = true
		}
	}

	if update {
		save_config(file, conf)
	}

	return conf, nil
}
