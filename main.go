package main

import (
	"encoding/csv"
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
		var lastFile string
		for {
			select {
			case event := <-watcher.Events:
				if event.Op.String() == "CREATE" {
					if lastFile != event.Name {
						f, err := os.Open(event.Name)
						if err != nil {
							log.Printf("Read error: %v %v\n", event.Name, err)
						}
						defer f.Close()

						var sell [][]string
						var buy [][]string
						lines, err := csv.NewReader(f).ReadAll()
						if err != nil {
							log.Println("Error reading lines", err)
						}

						for i, line := range lines {
							if i == 0 {
								//skip header
								continue
							}

							//only care about Jita 4-4 now
							if line[10] != "60003760" {
								continue
							}

							if line[7] == "True" {
								buy = append(buy, line)
							} else {
								sell = append(sell, line)
							}
						}

					}
					lastFile = event.Name
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
