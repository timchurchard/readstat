package pkg

import "time"

type KoboBook struct {
	ID    string
	Title string
	URL   string

	Finished        bool
	ProgressPercent int

	Parts map[string]KoboBookPart

	// IsBook true for book, false for Pocket
	IsBook bool
}

type KoboBookPart struct {
	WordCount int
}

// KoboEvent minimal event for reading start/stop/session
type KoboEvent struct {
	// BookID is the filename or Pocked ID
	BookID string

	EventType KoboEventType

	Time     time.Time
	Duration time.Time

	Words int
}

type KoboEventType string

const (
	KoboFilenamePrefix = "/mnt/onboard/"

	KoboTimeFmt    = "2006-01-02T15:05:06.000"
	koboTimeSample = "2023-12-19T11:42:00.000"

	ReadEvent   KoboEventType = "Read"
	FinishEvent KoboEventType = "Finish"
)

func (k KoboBook) TotalWords() int {
	result := 0

	for part := range k.Parts {
		result += k.Parts[part].WordCount
	}

	return result
}
