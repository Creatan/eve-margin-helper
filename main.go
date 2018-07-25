package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

func main() {
	logPath := os.Getenv("USERPROFILE") + "\\Documents\\EVE\\logs\\Marketlogs"
	fmt.Println("Listening to: ", logPath)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op.String() == "CREATE" || event.Op.String() == "WRITE" {
					log.Println("event:", event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}

		}
	}()

	err = watcher.Add(logPath)
	if err != nil {
		log.Fatal(err)
	}
	<-done

}
