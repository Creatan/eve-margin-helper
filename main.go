package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/atotto/clipboard"
	humanize "github.com/dustin/go-humanize"
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

func removeFiles(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.Remove(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}

	return nil

}
func main() {
	logPath := os.Getenv("USERPROFILE") + "\\Documents\\EVE\\logs\\Marketlogs"
	taxPer := flag.Float64("tax", 2.0, "Tax percentage after modifiers")
	feePer := flag.Float64("fee", 3.0, "Broker's fee percentage after modifiers")
	increment := flag.Float64("increment", 0.01, "Increment to overcut other buy orders")

	flag.Parse()
	//clear
	args := flag.Args()
	if len(args) != 0 {
		if args[0] == "clean" {
			err := removeFiles(logPath)
			if err != nil {
				log.Fatal("Error clearing log directory", err)
			}
			os.Exit(0)
		}
	}

	//convert percentages to decimals for calculations
	taxValue := *taxPer / 100
	feeValue := *feePer / 100

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
					if err != nil {
						log.Println("Error finding sell order value", err)
					}

					buyOrder, err := max(buy)
					if err != nil {
						log.Println("Error finding buy order value", err)
					}

					fees := (buyOrder * feeValue) + (sellOrder * feeValue)
					taxes := sellOrder * taxValue
					profit := sellOrder - fees - taxes - buyOrder
					profitPer := (profit / sellOrder) * 100

					log.Println("Sell", humanize.FormatFloat("# ###,##", sellOrder))
					log.Println("Buy", humanize.FormatFloat("# ###,##", buyOrder))
					log.Println("Fees", humanize.FormatFloat("# ###,##", fees))
					log.Println("Taxes", humanize.FormatFloat("# ###,##", taxes))
					log.Println("Profit", humanize.FormatFloat("# ###,##", profit))
					log.Printf("Profit %% %.2f\n", profitPer)
					err = clipboard.WriteAll(humanize.FormatFloat("####.##", buyOrder+*increment))
					if err != nil {
						log.Println("Error copying to clipboard", err)
					}
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
