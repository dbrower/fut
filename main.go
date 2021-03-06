package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/dbrower/fut/internal/fut"
)

type Config struct {
	Mysql          string
	Fedora         string
	StaticFilePath string
	TemplatePath   string
	Port           string
}

var (
	fedora *fut.RemoteFedora
	db     *fut.MysqlDB
)

func main() {
	config := Config{
		Mysql:          "",
		Fedora:         os.Getenv("FEDORA_PATH"),
		Port:           "8080",
		TemplatePath:   "./web/templates",
		StaticFilePath: "./web/static",
	}
	configFile := flag.String("config-file", "", "Configuration File to use")
	flag.Parse()
	if *configFile != "" {
		log.Println("Using config file", *configFile)
		if _, err := toml.DecodeFile(*configFile, &config); err != nil {
			log.Println(err)
			return
		}
	}

	if config.Fedora == "" {
		log.Println("FEDORA_PATH not set")
		return
	}
	fedora = fut.NewRemote(config.Fedora)
	fut.TargetFedora = fedora
	if config.Mysql != "" {
		var err error
		db, err = fut.NewMySQL(config.Mysql)
		if err != nil {
			log.Println(err)
			return
		}
		fut.Datasource = db
	}

	if config.TemplatePath != "" {
		err := fut.LoadTemplates(config.TemplatePath)
		if err != nil {
			log.Println(err)
		}
		fut.StaticFilePath = config.StaticFilePath
		// add routes
		h := fut.AddRoutes()
		go fut.BackgroundHarvester()
		http.ListenAndServe(":"+config.Port, h)
	}
}
