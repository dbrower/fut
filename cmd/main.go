package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/dbrower/fut"
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
		http.ListenAndServe(":"+config.Port, h)
	}
}

func DoHarvest() {
	t := time.Now()
	t = t.Add(-5 * 24 * time.Hour)

	f := fut.PrintItem
	c := make(chan fut.CurateItem, 10)
	if db != nil {
		f = func(item fut.CurateItem) error {
			c <- item
			return nil
		}
		go func() {
			for item := range c {
				err := db.IndexItem(item)
				if err != nil {
					log.Println(err)
				}
			}
		}()
	}
	log.Println(fedora, f)
	fut.HarvestCurateObjects(fedora, t, f)
	close(c)
}
