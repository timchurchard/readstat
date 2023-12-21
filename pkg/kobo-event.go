package pkg

import "time"

// KoboEvent minimal event for reading start/stop/session
type KoboEvent struct {
	// BookID is the filename or Pocket ID
	BookID string

	EventType KoboEventType

	Time            time.Time
	ReadingSessions []KoboEventReadingSession
}

type KoboEventReadingSession struct {
	UnixStart int
	UnixEnd   int

	Start time.Time
	End   time.Time
}

type KoboEventType string

func (t KoboEventType) String() string {
	return string(t)
}

const (
	KoboFilenamePrefix = "/mnt/onboard/"
	KoboPartSeparator  = "!!"

	KoboTimeFmt = "2006-01-02T15:04:05.000"
	// koboTimeSample = "2023-12-19T11:42:00.000"

	ReadEvent       KoboEventType = "Read"
	Progress25Event KoboEventType = "25%"
	Progress50Event KoboEventType = "50%"
	Progress75Event KoboEventType = "75%"
	FinishEvent     KoboEventType = "Finish"
)
