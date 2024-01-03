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

const (
	koboFilePartSeparator     = "!!"
	koboFilePartSeparatorAlt  = "#("
	koboDLPartSeparator       = "!OPS!"
	koboDLPartSeparatorLegacy = "!OEBPS!"
	koboDLPartEpub            = "!EPUB!"

	pocketMime = "application/x-kobo-html+pocket"

	downloadedType = "6" // downloadedType seems to cover 'downloaded' stuff from kobo/pocket

	// todo: not reliable
	// localfileType  = "9" // localfileType seems to cover local files e.g. all my side-loaded calibre (k)epubs
	// localfilePartType = "899" // localfilePartType seems to cover html files inside epubs
)

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
	stmt, _, err := k.conn.Prepare(`SELECT ContentID, BookID, ContentType, MimeType, Title, BookTitle, Attribution, ReadStatus, WordCount, ___PercentRead, ContentURL FROM content`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	index := map[string]int{}
	result := []KoboBook{}

	for stmt.Step() {
		cID := stmt.ColumnText(0)
		// bID := stmt.ColumnText(1)
		contentType := stmt.ColumnText(2)
		mimeType := stmt.ColumnText(3)
		title := stmt.ColumnText(4)
		// bookTitle := stmt.ColumnText(5)
		author := stmt.ColumnText(6)
		readStatus := stmt.ColumnInt(7)
		wordCount := stmt.ColumnInt(8)
		pcRead := stmt.ColumnInt(9)
		contentURL := stmt.ColumnText(10)

		// fmt.Printf("koboDatabase.Contents! cID=%s bID=%s contentType=%s mimeType=%s title=%s bookTitle=%s author=%s readStatus=%d wordCount=%d pcRead=%d contentURL=%s\n",
		//	cID, bID, contentType, mimeType, title, bookTitle, author, readStatus, wordCount, pcRead, contentURL)

		if contentType == downloadedType && mimeType == pocketMime {
			if _, exists := index[cID]; !exists {
				result = append(result, KoboBook{
					ID:              cID,
					Title:           title,
					Author:          author,
					URL:             contentURL,
					Finished:        readStatus == 1, // Pocket articles seem to have 0 or 1
					ProgressPercent: pcRead,
					Parts:           map[string]KoboBookPart{"0": {WordCount: wordCount}},
					IsBook:          false,
				})

				index[cID] = len(result) - 1
			}
		}

		if mimeType != pocketMime { // ignore downloaded pocket articles
			fn, pn := splitContentFilename(cID)

			if _, exists := index[fn]; !exists {
				result = append(result, KoboBook{
					ID:     fn,
					Parts:  map[string]KoboBookPart{},
					IsBook: true,
				})

				index[fn] = len(result) - 1
			}

			if pn == "" {
				result[index[fn]].Title = title
				result[index[fn]].Author = author
				result[index[fn]].URL = contentURL
				result[index[fn]].ProgressPercent = pcRead
			} else {
				if wordCount > 0 {
					if _, exists := result[index[fn]].Parts[pn]; !exists {
						result[index[fn]].Parts[pn] = KoboBookPart{
							WordCount: wordCount,
						}
					}
				}
			}
		}
	}
	if err = stmt.Err(); err != nil {
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
		eventProgress25  = 1012
		eventProgress50  = 1013
		eventProgress75  = 1014
		eventReadStart   = 1020
		eventReadEnd     = 1021
		eventFinished    = 80
		eventFinishedAlt = 5 // todo: I think this is 'ended reading session' e.g. switched book not finished book
		eventSession     = 46

		// minReadSessionSecs may need tweaking. Minimum reading session to include in stats
		minReadSessionSecs = 29

		extraDataReadingSeconds = "ExtraDataReadingSeconds"
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
		extraData := []byte{}
		colData := stmt.ColumnBlob(5, extraData)

		r := bytes.NewBuffer(colData)
		v, err := (&internal.QDataStreamReader{
			Reader:    r,
			ByteOrder: binary.BigEndian,
		}).ReadQStringQVariantAssociative()
		if err != nil {
			// Ignore EOF & unimplemented errors when decoding extra data
			if !errors.Is(err, io.EOF) && err.Error() != "unimplemented type 20" {
				fmt.Println(err)
			}
		}

		// first := stmt.ColumnText(1)
		last := stmt.ColumnText(2)
		// count := stmt.ColumnInt(3)
		lastTime, _ := time.Parse(KoboTimeFmt, last)

		// Try to get filename from cID
		cID := stmt.ColumnText(4)
		fn, _ := splitContentFilename(cID)
		if fn == "" {
			continue
		}

		/*if strings.Contains(fn, "Fight") {
			fmt.Printf("DEBUG! %d / %s / %s / %v\n", stmt.ColumnInt(0), fn, last, v)
		}*/

		if strings.HasSuffix(fn, ".png") {
			// Skip image files (koreader.png for example)
			continue
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

		case eventSession:
			// For pocket we only get eventSession
			contentType := ""
			for key, val := range v {
				if key == "ContentType" {
					switch val.(type) {
					case string:
						contentType = val.(string)
					case []uint8:
						contentType = string(val.([]uint8))
					}
				}
			}

			if contentType == pocketMime {
				// Pocket mime
				// fmt.Printf("DEBUG! %d / %s / %s / %s / %d / %v\n", stmt.ColumnInt(0), cID, first, last, count, v)
				if v[extraDataReadingSeconds] != nil {
					// We got a non-nil reading seconds. (Finished & percent is stored in content table for pocket)
					secondsRead := int(v[extraDataReadingSeconds].(int32))

					if secondsRead > minReadSessionSecs {
						startUnix := int(lastTime.Unix())

						sessions := []KoboEventReadingSession{
							{
								UnixStart: startUnix,
								UnixEnd:   startUnix + secondsRead,
								Start:     time.Unix(int64(startUnix), 0),
								End:       time.Unix(int64(startUnix+secondsRead), 0),
							},
						}
						result = append(result, KoboEvent{BookID: fn, EventType: ReadEvent, Time: time.Time{}, ReadingSessions: sessions})
					}
				}
			}

		default:
			// fmt.Printf("DEBUG! %d / %s / %s / %s / %d / %v\n", stmt.ColumnInt(0), fn, first, last, count, v)
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
