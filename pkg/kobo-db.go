//go:generate mockgen -package pkg -destination kobo-db_mock.go -source kobo-db.go
package pkg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/timchurchard/readstat/internal"
)

type KoboDatabase interface {
	Device() (string, string)

	Contents() ([]KoboBook, error)
	Events() ([]KoboEvent, error)

	Close() error
}

type koboDatabase struct {
	fn   string
	conn *sqlite3.Conn

	// device contains the first value from the .kobo/version file (model + serial)
	device string
	model  string
}

func NewKoboDatabase(fn string) (KoboDatabase, error) {
	// todo read-only ! conn, err := sqlite3.OpenFlags(fn, sqlite3.OPEN_READONLY)
	conn, err := sqlite3.Open(fn)
	if err != nil {
		return nil, err
	}

	device, err := getDevice(fn)
	if err != nil {
		return nil, err
	}

	return koboDatabase{
		fn:     fn,
		conn:   conn,
		device: device,
		model:  getModel(device),
	}, nil
}

func (k koboDatabase) Device() (string, string) {
	return k.device, k.model
}

func (k koboDatabase) Contents() ([]KoboBook, error) {
	stmt, _, err := k.conn.Prepare(`SELECT ContentID, BookID, BookTitle, ReadStatus, WordCount, BookmarkWordOffset, ___PercentRead, TimeSpentReading FROM content`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	index := map[string]int{}
	result := []KoboBook{}

	for stmt.Step() {
		if strings.Contains(stmt.ColumnText(1), KoboFilenamePrefix) {
			// Found a local file like the Calibre loaded .epub files I read. TODO kobo downloaded content. TODO Pocket content. TODO other?
			fn := cleanContentFilename(stmt.ColumnText(1))

			idx, exists := index[fn]
			if !exists {
				result = append(result, KoboBook{
					ID:              fn,
					Title:           stmt.ColumnText(2),
					URL:             "", // todo: Pocket
					Finished:        false,
					ProgressPercent: 0,
					Parts:           map[string]KoboBookPart{},
					IsBook:          true,
				})

				index[fn] = len(result) - 1
				idx = index[fn]
			}

			if strings.Contains(stmt.ColumnText(0), KoboPartSeparator) {
				// Book part, used for word count
				part := getPartOffFilename(stmt.ColumnText(0))

				wordCount := 0
				if stmt.ColumnInt(4) > 0 {
					wordCount = stmt.ColumnInt(4)
				}

				if _, exists = result[idx].Parts[part]; !exists {
					result[idx].Parts[part] = KoboBookPart{
						WordCount: wordCount,
					}
				}
			} else {
				// Book, update fields
				// Record rough percentage (25, 50, 75, 100)
				result[idx].ProgressPercent = stmt.ColumnInt(6)

				// If ReadStatus is 2 then we'll set IsFinished
				if stmt.ColumnInt(3) == 2 {
					result[idx].Finished = true
				}
			}
		}
	}
	if err := stmt.Err(); err != nil {
		return nil, err
	}

	err = stmt.Close()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (k koboDatabase) Events() ([]KoboEvent, error) {
	const (
		eventProgress25 = 1012
		eventProgress50 = 1013
		eventProgress75 = 1014
		eventReadStart  = 1020
		eventReadEnd    = 1021
		eventFinished   = 80

		// minReadSessionSecs may need tweaking. Minimum reading session to include in stats
		minReadSessionSecs = 30
	)

	stmt, _, err := k.conn.Prepare(`SELECT EventType, FirstOccurrence, LastOccurrence, EventCount, ContentID, ExtraData FROM Event`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result := []KoboEvent{} // todo: I think events are unique by (type + content) so not handling duplicates now!

	// startTimes holds the list of times from the 1020 event. We rely on startTimes/endTimes being same length and we'll pair them 0=0 etc
	startTimes := map[string][]uint32{}
	endTimes := map[string][]uint32{}

	for stmt.Step() {
		if strings.Contains(stmt.ColumnText(4), KoboFilenamePrefix) {
			extraData := []byte{}
			colData := stmt.ColumnBlob(5, extraData)

			r := bytes.NewBuffer(colData)
			v, err := (&internal.QDataStreamReader{
				Reader:    r,
				ByteOrder: binary.BigEndian,
			}).ReadQStringQVariantAssociative()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Println(err)
				}
			}

			// Found a local file like the Calibre loaded .epub files I read. TODO kobo downloaded content. TODO Pocket content. TODO other?
			fn := cleanContentFilename(stmt.ColumnText(4))

			// first := stmt.ColumnText(1)
			last := stmt.ColumnText(2)
			// count := stmt.ColumnInt(3)

			lastTime, err := time.Parse(KoboTimeFmt, last)
			if err != nil {
				fmt.Println(err)
			}

			switch stmt.ColumnInt(0) {
			case eventProgress25:
				result = append(result, KoboEvent{BookID: fn, EventType: Progress25Event, Time: lastTime})

			case eventProgress50:
				result = append(result, KoboEvent{BookID: fn, EventType: Progress50Event, Time: lastTime})

			case eventProgress75:
				result = append(result, KoboEvent{BookID: fn, EventType: Progress75Event, Time: lastTime})

			case eventReadStart:
				data := v["eventTimestamps"].([]interface{})
				startTimes[fn] = make([]uint32, len(data))
				for idx := range data {
					startTimes[fn][idx] = data[idx].(uint32)
				}

			case eventReadEnd:
				data := v["eventTimestamps"].([]interface{})
				endTimes[fn] = make([]uint32, len(data))
				for idx := range data {
					endTimes[fn][idx] = data[idx].(uint32)
				}

			case eventFinished:
				result = append(result, KoboEvent{BookID: fn, EventType: FinishEvent, Time: lastTime})

			default:
				// fmt.Printf("DEBUG! %d / %s / %s / %s / %d / %v\n", stmt.ColumnInt(0), fn, first, last, count, v)
			}
		}
	}
	if err := stmt.Err(); err != nil {
		return nil, err
	}

	err = stmt.Close()
	if err != nil {
		return nil, err
	}

	// Process all the startTimes/endTimes to make reading events
	for fn := range startTimes {
		if _, exists := endTimes[fn]; !exists {
			continue
		}

		sessions := []KoboEventReadingSession{}

		for idx := range startTimes[fn] {
			if endTimes[fn][idx]-startTimes[fn][idx] < minReadSessionSecs {
				continue
			}

			sessions = append(sessions, KoboEventReadingSession{
				UnixStart: int(startTimes[fn][idx]),
				UnixEnd:   int(endTimes[fn][idx]),
				Start:     time.Unix(int64(startTimes[fn][idx]), 0),
				End:       time.Unix(int64(endTimes[fn][idx]), 0),
			})
		}

		result = append(result, KoboEvent{
			BookID:          fn,
			EventType:       ReadEvent,
			Time:            time.Time{},
			ReadingSessions: sessions,
		})
	}

	return result, nil
}

func (k koboDatabase) Close() error {
	return k.conn.Close()
}
