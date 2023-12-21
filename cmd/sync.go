package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/timchurchard/readstat/pkg"
)

func Sync(out io.Writer) int {
	const (
		defaultEmpty = ""

		usageDatabasePath = "Path to /media/kobo/.kobo/KoboReader.sqlite"
	)
	var (
		databaseFn string
	)

	flag.StringVar(&databaseFn, "database", defaultEmpty, usageDatabasePath)
	flag.StringVar(&databaseFn, "d", defaultEmpty, usageDatabasePath)

	flag.Usage = func() {
		fmt.Fprintf(out, "Usage of %s %s:\n", os.Args[0], os.Args[1])

		flag.PrintDefaults()
	}

	flag.Parse()

	db, err := pkg.NewKoboDatabase(databaseFn)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	// db.Events()
	db.Contents()

	return 0
}
