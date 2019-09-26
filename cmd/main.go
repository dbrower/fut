package main

import (
	"log"
	"os"
	"time"

	"github.com/dbrower/fut"
)

func main() {
	fpath := os.Getenv("FEDORA_PATH")
	if fpath == "" {
		log.Println("FEDORA_PATH not set")
		return
	}
	fedora := fut.NewRemote(fpath)
	t := time.Now()
	t = t.Add(-5 * 24 * time.Hour)

	fut.HarvestCurateObjects(fedora, t, fut.PrintItem)
}
