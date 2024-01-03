//go:generate mockgen -package pkg -destination stats_mock.go -source stats.go
package pkg

import (
	"fmt"
	"strings"
	"time"

	"github.com/timchurchard/readstat/internal"
)

type Stats interface {
	BooksFinishedYear() []StatsBook
	BooksFinishedMonth(month int) []StatsBook

	ArticlesFinishedYear() []StatsBook
	ArticlesFinishedMonth(month int) []StatsBook

	WordsFinishedYear() int
	WordsFinishedMonth(month int) int

	BooksSecondsReadYear() int
	BooksSecondsReadMonth(month int) int

	ArticlesSecondsReadYear() int
	ArticlesSecondsReadMonth(month int) int
}

type YearStats struct {
	Year   int                `json:"year"`
	Months map[int]MonthStats `json:"months"`
}

type MonthStats struct {
	FinishedBooks    []StatsBook `json:"finished"`
	FinishedArticles []StatsBook `json:"finished_articles"`

	Books    []StatsBook `json:"progressed"`
	Articles []StatsBook `json:"progressed_articles"`
}

func (m *MonthStats) AddFinishedBook(sb StatsBook) {
	found := false
	for idx := range m.FinishedBooks {
		if m.FinishedBooks[idx].BookID == sb.BookID {
			found = true
			break
		}
	}

	if !found {
		m.FinishedBooks = append(m.FinishedBooks, sb)
	}
}

func (m *MonthStats) AddFinishedArticle(sb StatsBook) {
	found := false
	for idx := range m.FinishedArticles {
		if m.FinishedArticles[idx].BookID == sb.BookID {
			found = true
			break
		}
	}

	if !found {
		m.FinishedArticles = append(m.FinishedArticles, sb)
	}
}

func (m *MonthStats) AddBook(sb StatsBook) {
	found := false
	for idx := range m.Books {
		if m.Books[idx].BookID == sb.BookID {
			found = true
			break
		}
	}

	if !found {
		m.Books = append(m.Books, sb)
	}
}

func (m *MonthStats) AddArticle(sb StatsBook) {
	found := false
	for idx := range m.Articles {
		if m.Articles[idx].BookID == sb.BookID {
			found = true
			break
		}
	}

	if !found {
		m.Articles = append(m.Articles, sb)
	}
}

func (m *MonthStats) AddReading(bookID string, startTime string, duration int) {
	read := StatsRead{
		Time:     startTime,
		Duration: duration,
	}

	for idx := range m.Books {
		if m.Books[idx].BookID == bookID {
			m.Books[idx].Reads = append(m.Books[idx].Reads, read)
		}
	}
	for idx := range m.Articles {
		if m.Articles[idx].BookID == bookID {
			m.Articles[idx].Reads = append(m.Articles[idx].Reads, read)
		}
	}
}

type StatsBook struct {
	BookID string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	URL    string `json:"url"`

	Words int `json:"words"`

	IsBook     bool `json:"is_book"`
	IsFinished bool `json:"is_finished"`

	Reads []StatsRead `json:"reads"`
}

func (b StatsBook) ReadSeconds() int {
	result := 0

	for idx := range b.Reads {
		result += b.Reads[idx].Duration
	}

	return result
}

type StatsRead struct {
	Time     string `json:"time"`
	Duration int    `json:"duration"`
}

// NewStatsForYear take a Storage and calculate the stats for a given year (4 character e.g. 2023 int)
func NewStatsForYear(storage internal.Storage, year int) YearStats {
	months := make([]MonthStats, 13) // We'll use 1..12 as the range
	for idx := 1; idx <= 12; idx++ {
		months[idx].FinishedBooks = []StatsBook{}
		months[idx].Books = []StatsBook{}
	}

	// Books in months
	for cIdx := range storage.Contents {
		if strings.Contains(cIdx, "Nevere") {
			fmt.Println(cIdx)
		}

		if storage.Contents[cIdx].IsBook {
			// First pass finished books
			for eIdx := range storage.Events[cIdx] {
				eventTime, _ := time.Parse(internal.StorageTimeFmt, storage.Events[cIdx][eIdx].Time)

				inYear, monthNo := isInYearAndMonth(year, eventTime)
				if inYear {
					switch storage.Events[cIdx][eIdx].EventName {
					case "Finish":
						months[monthNo].AddFinishedBook(StatsBook{
							BookID:     cIdx,
							Title:      storage.Contents[cIdx].Title,
							Author:     storage.Contents[cIdx].Author,
							Words:      storage.Contents[cIdx].Words,
							IsFinished: true,
							Reads:      nil,
						})
					}
				}
			}

			// Second pass any read book ensure exists in the right months
			for eIdx := range storage.Events[cIdx] {
				eventTime, _ := time.Parse(internal.StorageTimeFmt, storage.Events[cIdx][eIdx].Time)

				inYear, monthNo := isInYearAndMonth(year, eventTime)
				if inYear {
					switch storage.Events[cIdx][eIdx].EventName {
					case "Read":
						months[monthNo].AddBook(StatsBook{
							BookID:     cIdx,
							Title:      storage.Contents[cIdx].Title,
							Author:     storage.Contents[cIdx].Author,
							Words:      storage.Contents[cIdx].Words,
							IsFinished: false,
							Reads:      []StatsRead{},
						})
					}
				}
			}

			// Third pass reading sessions!
			for eIdx := range storage.Events[cIdx] {
				eventTime, _ := time.Parse(internal.StorageTimeFmt, storage.Events[cIdx][eIdx].Time)

				inYear, monthNo := isInYearAndMonth(year, eventTime)
				if inYear {
					switch storage.Events[cIdx][eIdx].EventName {
					case "Read":
						months[monthNo].AddReading(cIdx, storage.Events[cIdx][eIdx].Time, storage.Events[cIdx][eIdx].Duration)
					}
				}
			}
		} else {
			for eIdx := range storage.Events[cIdx] {
				eventTime, _ := time.Parse(internal.StorageTimeFmt, storage.Events[cIdx][eIdx].Time)

				inYear, monthNo := isInYearAndMonth(year, eventTime)
				if inYear {
					switch storage.Events[cIdx][eIdx].EventName {
					case "Read":
						sb := StatsBook{
							BookID:     cIdx,
							Title:      storage.Contents[cIdx].Title,
							Author:     storage.Contents[cIdx].Author,
							URL:        storage.Contents[cIdx].URL,
							Words:      storage.Contents[cIdx].Words,
							IsBook:     false,
							IsFinished: storage.Contents[cIdx].IsFinished,
							Reads:      []StatsRead{},
						}

						if storage.Contents[cIdx].IsFinished {
							months[monthNo].AddFinishedArticle(sb)
						}

						months[monthNo].AddArticle(sb)
						months[monthNo].AddReading(cIdx, storage.Events[cIdx][eIdx].Time, storage.Events[cIdx][eIdx].Duration)
					}
				}
			}
		}
	}

	// build the result!
	result := YearStats{
		Year:   year,
		Months: map[int]MonthStats{},
	}

	for idx := 1; idx <= 12; idx++ {
		result.Months[idx] = months[idx]
	}

	return result
}

func (y YearStats) BooksFinishedYear() []StatsBook {
	result := []StatsBook{}

	for idx := range y.Months {
		result = append(result, y.BooksFinishedMonth(idx)...)
	}

	return result
}

func (y YearStats) BooksFinishedMonth(month int) []StatsBook {
	return y.Months[month].FinishedBooks
}

func (y YearStats) ArticlesFinishedYear() []StatsBook {
	result := []StatsBook{}

	for idx := range y.Months {
		result = append(result, y.ArticlesFinishedMonth(idx)...)
	}

	return result
}

func (y YearStats) ArticlesFinishedMonth(month int) []StatsBook {
	return y.Months[month].FinishedArticles
}

func (y YearStats) WordsFinishedYear() int {
	result := 0

	for idx := range y.Months {
		result += y.WordsFinishedMonth(idx)
	}

	return result
}

func (y YearStats) WordsFinishedMonth(month int) int {
	result := 0

	for _, book := range append(y.BooksFinishedMonth(month), y.ArticlesFinishedMonth(month)...) {
		result += book.Words
	}

	return result
}

func (y YearStats) BooksSecondsReadYear() int {
	result := 0

	for idx := range y.Months {
		result += y.BooksSecondsReadMonth(idx)
	}

	return result
}

func (y YearStats) BooksSecondsReadMonth(month int) int {
	result := 0

	for _, book := range y.Months[month].Books {
		result += book.ReadSeconds()
	}

	return result
}

func (y YearStats) ArticlesSecondsReadYear() int {
	result := 0

	for idx := range y.Months {
		result += y.ArticlesSecondsReadMonth(idx)
	}

	return result
}

func (y YearStats) ArticlesSecondsReadMonth(month int) int {
	result := 0

	for _, book := range y.Months[month].Articles {
		result += book.ReadSeconds()
	}

	return result
}

// isInYearAndMonth todo refactor
func isInYearAndMonth(year int, eventTime time.Time) (bool, int) {
	yearStart, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-01-01", year))
	Feb, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-02-01", year))
	Mar, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-03-01", year))
	Apr, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-04-01", year))
	May, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-05-01", year))
	Jun, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-06-01", year))
	Jul, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-07-01", year))
	Aug, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-08-01", year))
	Sep, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-09-01", year))
	Oct, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-10-01", year))
	Nov, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-11-01", year))
	Dec, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-12-01", year))
	yearEnd, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-01-01", year+1))

	if inTimeSpan(yearStart, yearEnd, eventTime) {
		month := -1

		if inTimeSpan(yearStart, Feb, eventTime) {
			month = 1
		}
		if inTimeSpan(Feb, Mar, eventTime) {
			month = 2
		}
		if inTimeSpan(Mar, Apr, eventTime) {
			month = 3
		}
		if inTimeSpan(Apr, May, eventTime) {
			month = 4
		}
		if inTimeSpan(May, Jun, eventTime) {
			month = 5
		}
		if inTimeSpan(Jun, Jul, eventTime) {
			month = 6
		}
		if inTimeSpan(Jul, Aug, eventTime) {
			month = 7
		}
		if inTimeSpan(Aug, Sep, eventTime) {
			month = 8
		}
		if inTimeSpan(Sep, Oct, eventTime) {
			month = 9
		}
		if inTimeSpan(Oct, Nov, eventTime) {
			month = 10
		}
		if inTimeSpan(Nov, Dec, eventTime) {
			month = 11
		}
		if inTimeSpan(Dec, yearEnd, eventTime) {
			month = 12
		}

		return true, month
	}

	return false, -1
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
