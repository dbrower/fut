package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/dbrower/fut"
)

type Config struct {
	Mysql  string
	Fedora string
}

func main() {
	config := Config{
		Mysql:  "",
		Fedora: os.Getenv("FEDORA_PATH"),
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
	fedora := fut.NewRemote(config.Fedora)
	t := time.Now()
	t = t.Add(-5 * 24 * time.Hour)

	f := fut.PrintItem
	c := make(chan fut.CurateItem, 10)
	var db *fut.MysqlDB
	if config.Mysql != "" {
		var err error
		db, err = fut.NewMySQL(config.Mysql)
		if err != nil {
			log.Println(err)
			return
		}
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
	//fut.HarvestCurateObjects(fedora, t, f)
	close(c)

	if db != nil {
		v, err := db.FindAllRange(0, 10)
		if err != nil {
			log.Println(err)
		}
		for _, vv := range v {
			log.Println(vv)
		}
	}
}
