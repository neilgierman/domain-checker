package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Database Database `json:"database"`
	Queue Queue `json:"queue"`
}

type Database struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Database string `json:"database"`
}

type Queue struct {
	Host string `json:"host"`
	User string `json:"user"`
	Password string `json:"password"`
	Port string `json:"port"`
}

func (a *App) loadConfig() {
	a.appCfg = &Config{}
	f, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	byteValue, _ := ioutil.ReadAll(f)
	err = json.Unmarshal(byteValue, a.appCfg)
	if err != nil {
		log.Fatal(err)
	}

	a.dbConfig = &DBConfig{}
}
