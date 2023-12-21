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
)

// Stats command reads local storage and produces stats
func Stats(out io.Writer) int {
	const (
		// defaultEmpty   = ""
		defaultStorage = "./readstat.json"

		usageStoragePath = "Path to local storage default: " + defaultStorage
	)
	var (
		storageFn string
	)

	// TODO year - what year to make stats for
	// TODO mode - output text summary ? html ? etc

	flag.StringVar(&storageFn, "storage", defaultStorage, usageStoragePath)
	flag.StringVar(&storageFn, "s", defaultStorage, usageStoragePath)

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

	// TC Stats proof of concept ?

	yearStart, _ := time.Parse("2006-01-02", "2023-01-01")
	yearEnd, _ := time.Parse("2006-01-02", "2024-01-01")

	// numFinished All books finished in 2023
	numFinished := 0
	numFinishedWords := 0

	totalReadingSeconds := 0

	for cIdx := range storage.Contents {
		for eIdx := range storage.Events[cIdx] {
			eventTime, _ := time.Parse(internal.StorageTimeFmt, storage.Events[cIdx][eIdx].Time)
			if inTimeSpan(yearStart, yearEnd, eventTime) {

				switch storage.Events[cIdx][eIdx].EventName {
				case "Finish":
					numFinished += 1
					numFinishedWords += storage.Contents[cIdx].Words
					break

				case "Read":
					totalReadingSeconds += storage.Events[cIdx][eIdx].Duration
				}
			}
		}
	}

	totalReadingDuration, err := time.ParseDuration(fmt.Sprintf("%ds", totalReadingSeconds))
	if err != nil {
		panic(err)
	}

	fmt.Println("In 2023 / Tim")
	fmt.Printf("Finished books\t\t\t: %d\n", numFinished)
	fmt.Printf("Finish books (words)\t\t: %s\n", humanizeInt(numFinishedWords))
	fmt.Printf("Time spent reading books\t: %s", humanizeDuration(totalReadingDuration))

	return 0
}

// inTimeSpan check time in range
// From: https://stackoverflow.com/a/55093788
func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
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
