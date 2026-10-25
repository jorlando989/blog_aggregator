package config

import (
	"os"
	"encoding/json"
	"log"
)

type Config struct {
	DbUrl            string
	CurrentUserName string
}

const configFileName = ".gatorconfig.json"

func getComputedPath() string {
	homeDir, _ := os.UserHomeDir()
	return homeDir + "/" + configFileName
}


func Read() Config {
	content, err := os.ReadFile(getComputedPath())
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		log.Fatal(err)
	}

	return Config{
		DbUrl: config.DbUrl,
		CurrentUserName: config.CurrentUserName,
	}
}

func (c *Config) SetUser(user string) {
	//write config struct to json file
	c.CurrentUserName = user

	content, err := json.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}
	err2 := os.WriteFile(getComputedPath(), content, 0644)
	if err2 != nil {
		log.Fatalf("Failed to write file: %v", err2)
	}
}