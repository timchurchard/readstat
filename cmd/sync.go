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

	defer storage.Save()

	storage.AddDevice(db.Device())

	for cIdx := range contents {
		storage.AddContent(contents[cIdx].ID, contents[cIdx].Title, contents[cIdx].URL, contents[cIdx].TotalWords(), contents[cIdx].IsBook)
	}

	for eIdx := range events {
		if events[eIdx].EventType == pkg.ReadEvent {
			for sIdx := range events[eIdx].ReadingSessions {
				durationSecs := events[eIdx].ReadingSessions[sIdx].UnixEnd - events[eIdx].ReadingSessions[sIdx].UnixStart
				storage.AddEvent(events[eIdx].BookID, events[eIdx].EventType.String(), events[eIdx].ReadingSessions[sIdx].Start, durationSecs)
			}
		} else {
			storage.AddEvent(events[eIdx].BookID, events[eIdx].EventType.String(), events[eIdx].Time, 0)
		}
	}

	return 0
}
