package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/timchurchard/readstat/internal"
	"github.com/timchurchard/readstat/pkg"
)

// Sync command reads a Kobo database and creates/updates local storage
func Sync(out io.Writer) int {
	const (
		defaultEmpty   = ""
		defaultStorage = "./readstat.json"

		usageDatabasePath = "Path to /media/kobo/.kobo/KoboReader.sqlite"
		usageStoragePath  = "Path to local storage default: " + defaultStorage
	)
	var (
		databaseFn string
		storageFn  string
	)

	flag.StringVar(&databaseFn, "database", defaultEmpty, usageDatabasePath)
	flag.StringVar(&databaseFn, "d", defaultEmpty, usageDatabasePath)

	flag.StringVar(&storageFn, "storage", defaultStorage, usageStoragePath)
	flag.StringVar(&storageFn, "s", defaultStorage, usageStoragePath)

	flag.Usage = func() {
		fmt.Fprintf(out, "Usage of %s %s:\n", os.Args[0], os.Args[1])

		flag.PrintDefaults()
	}

	flag.Parse()

	if databaseFn == "" {
		fmt.Println("-d or --database /path/to/KoboReader.sqlite is required.")
		return 1
	}

	// Read data from Kobo DB
	db, err := pkg.NewKoboDatabase(databaseFn)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	contents, err := db.Contents()
	if err != nil {
		panic(err)
	}

	events, err := db.Events()
	if err != nil {
		panic(err)
	}

	// Create/Update Storage
	storage, err := internal.OpenStorage(storageFn)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := storage.Save(); err != nil {
			fmt.Printf("Error saving: %v\n", err)
		}
	}()

	device, model := db.Device()
	storage.AddDevice(device, model)

	for cIdx := range contents {
		storage.AddContent(contents[cIdx].ID, contents[cIdx].Title, contents[cIdx].Author, contents[cIdx].URL, contents[cIdx].TotalWords(), contents[cIdx].IsBook, contents[cIdx].Finished)
	}

	for eIdx := range events {
		if events[eIdx].EventType == pkg.ReadEvent {
			for sIdx := range events[eIdx].ReadingSessions {
				durationSecs := events[eIdx].ReadingSessions[sIdx].UnixEnd - events[eIdx].ReadingSessions[sIdx].UnixStart
				storage.AddEvent(events[eIdx].BookID, device, events[eIdx].EventType.String(), events[eIdx].ReadingSessions[sIdx].Start, durationSecs)
			}
		} else {
			storage.AddEvent(events[eIdx].BookID, device, events[eIdx].EventType.String(), events[eIdx].Time, 0)
		}
	}

	return 0
}
