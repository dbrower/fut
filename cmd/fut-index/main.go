package main

import (
	"bufio"
	"flag"
	"log"
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

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		pid := scanner.Text()
		item, err := fut.FetchOneCurateObject(fedora, pid)
		if err != nil {
			log.Println(pid, err)
			continue
		}
		err = fut.Datasource.IndexItem(item)
		if err != nil {
			log.Println(pid, err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
