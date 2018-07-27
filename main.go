package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

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
						log.Println("Found new file:", event.Name)
						f, err := os.Open(event.Name)
						if err != nil {
							log.Printf("Read error: %v %v\n", event.Name, err)
						}
						defer f.Close()

						var sell []float64
						var buy []float64
						reader := csv.NewReader(bufio.NewReader(f))
						reader.FieldsPerRecord = 15
						lineCount := 0
						for {
							if lineCount == 0 {
								//skip header
								lineCount++
								continue
							}

							line, err := reader.Read()
							if err == io.EOF {
								break
							}
							if len(line) == 0 {
								continue
							}
							if len(line) < 10 {
								log.Println("faulty", line)
								continue
							}
							if err != nil {
								log.Println("Error reading line", err)
							}

							lineCount++

							//only care about Jita 4-4 now
							if line[10] != "60003760" {
								continue
							}

							price, err := strconv.ParseFloat(line[0], 64)
							if err != nil {
								log.Println("Error parsing price", line[0], err)
							}

							if line[7] == "True" {
								buy = append(buy, price)
							} else {
								sell = append(sell, price)
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
