package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jmlattanzi/itaic-backend/itaic/models"
)

var config models.Config

// LoadConfigurationFile ...Loads a configuration file
func LoadConfigurationFile(filename string) models.Config {
	fmt.Println("[-] Loading configuration....")
	// Open the file and defer closing it until the function is done
	configFile, err := os.Open(filename)
	defer configFile.Close()
	if err != nil {
		log.Fatal("[!] Error loading configuration: ", err)
	}

	// decode the json and store it in config
	json.NewDecoder(configFile).Decode(&config)
	fmt.Println("[+] Configuration loaded successfully")
	return config
}
