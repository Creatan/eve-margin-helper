package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/fsnotify/fsnotify"
)

func max(values []float64) (max float64, err error) {
	if len(values) == 0 {
		return 0, errors.New("Slice is empty")
	}

	max = values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max, nil
}

func min(values []float64) (min float64, err error) {
	if len(values) == 0 {
		return 0, errors.New("Slice is empty")
	}

	min = values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min, nil
}

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
				if event.Op.String() == "WRITE" {
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
					//find max buy and min sell
					sellOrder, err := min(sell)
					log.Println("Sell", sellOrder)
					if err != nil {
						log.Println("Error finding sell order value", err)
					}

					buyOrder, err := max(buy)
					if err != nil {
						log.Println("Error finding buy order value", err)
					}

					log.Println("Buy", buyOrder)
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
