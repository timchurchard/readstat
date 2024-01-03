package cmd

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/timchurchard/readstat/internal"
	"github.com/timchurchard/readstat/pkg"
)

// Stats command reads local storage and produces stats
func Stats(out io.Writer) int {
	const (
		// defaultEmpty   = ""
		defaultStorage = "./readstat.json"
		defaultYear    = 2023

		usageStoragePath = "Path to local storage default: " + defaultStorage
		usageYear        = "Year to generate stats for (default 2023)"
	)
	var (
		storageFn string
		year      int
	)

	// TODO mode - output text summary ? html ? etc

	flag.StringVar(&storageFn, "storage", defaultStorage, usageStoragePath)
	flag.StringVar(&storageFn, "s", defaultStorage, usageStoragePath)

	flag.IntVar(&year, "year", defaultYear, usageYear)
	flag.IntVar(&year, "y", defaultYear, usageYear)

	flag.Usage = func() {
		fmt.Fprintf(out, "Usage of %s %s:\n", os.Args[0], os.Args[1])

		flag.PrintDefaults()
	}

	flag.Parse()

	// Read Storage
	storage, err := internal.OpenStorage(storageFn)
	if err != nil {
		panic(err)
	}

	stats := pkg.NewStatsForYear(storage, year)

	booksReadSeconds := stats.BooksSecondsReadYear()
	booksReadDuration, _ := time.ParseDuration(fmt.Sprintf("%ds", booksReadSeconds))
	articlesReadSeconds := stats.ArticlesSecondsReadYear()
	articlesReadDuration, _ := time.ParseDuration(fmt.Sprintf("%ds", articlesReadSeconds))

	totalReadSeconds := booksReadSeconds + articlesReadSeconds
	totalReadDuration, _ := time.ParseDuration(fmt.Sprintf("%ds", totalReadSeconds))

	fmt.Println("In 2023 / Tim")
	fmt.Printf("Finished books\t\t\t: %d\n", len(stats.BooksFinishedYear()))
	fmt.Printf("Finished articles\t\t: %d\n", len(stats.ArticlesFinishedYear()))
	fmt.Printf("Total finished words\t\t: %s\n", humanizeInt(stats.WordsFinishedYear()))
	fmt.Printf("Time reading books\t\t: %s\n", humanizeDuration(booksReadDuration))
	fmt.Printf("Time reading articles\t\t: %s\n", humanizeDuration(articlesReadDuration))
	fmt.Printf("Total time reading\t\t: %s\n", humanizeDuration(totalReadDuration))

	fmt.Println("\n----------")

	months := []string{"", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	for idx := 1; idx <= 12; idx++ {
		monthBookReadDuration, _ := time.ParseDuration(fmt.Sprintf("%ds", stats.BooksSecondsReadMonth(idx)))
		monthArticleReadDuration, _ := time.ParseDuration(fmt.Sprintf("%ds", stats.ArticlesSecondsReadMonth(idx)))

		fmt.Printf("\n%s %d - Finished books: %d, articles: %d, time spend reading books: %s and articles: %s\n", months[idx], year, len(stats.BooksFinishedMonth(idx)), len(stats.ArticlesFinishedMonth(idx)), humanizeDuration(monthBookReadDuration), humanizeDuration(monthArticleReadDuration))

		for _, finishedBook := range stats.BooksFinishedMonth(idx) {
			fmt.Printf("\t finished book: %s - %s\n", finishedBook.Title, finishedBook.Author)
		}

		/*for _, finishedArticle := range stats.ArticlesFinishedMonth(idx) {
			fmt.Printf("\t finished article: %s - %s (%s)\n", finishedArticle.Title, finishedArticle.Author, finishedArticle.URL)
		}*/
	}

	return 0
}

// humanizeDuration humanizes time.Duration output to a meaningful value,
// golang's default “time.Duration“ output is badly formatted and unreadable.
// From: https://gist.github.com/harshavardhana/327e0577c4fed9211f65
func humanizeDuration(duration time.Duration) string {
	if duration.Seconds() < 60.0 {
		return fmt.Sprintf("%d seconds", int64(duration.Seconds()))
	}
	if duration.Minutes() < 60.0 {
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%d minutes %d seconds", int64(duration.Minutes()), int64(remainingSeconds))
	}
	if duration.Hours() < 24.0 {
		remainingMinutes := math.Mod(duration.Minutes(), 60)
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%d hours %d minutes %d seconds",
			int64(duration.Hours()), int64(remainingMinutes), int64(remainingSeconds))
	}
	remainingHours := math.Mod(duration.Hours(), 24)
	remainingMinutes := math.Mod(duration.Minutes(), 60)
	remainingSeconds := math.Mod(duration.Seconds(), 60)
	return fmt.Sprintf("%d days %d hours %d minutes %d seconds",
		int64(duration.Hours()/24), int64(remainingHours),
		int64(remainingMinutes), int64(remainingSeconds))
}

// humanizeInt
// Based on https://github.com/dustin/go-humanize/blob/v1.0.1/comma.go#L15
func humanizeInt(num int) string {
	parts := []string{"", "", "", "", "", "", ""}
	j := len(parts) - 1

	for num > 999 {
		parts[j] = strconv.FormatInt(int64(num%1000), 10)
		num = num / 1000
		j--
	}

	parts[j] = strconv.Itoa(int(num))
	return strings.Join(parts[j:], ",")
}
