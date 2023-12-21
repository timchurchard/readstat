//go:generate mockgen -package pkg -destination kobo-db_mock.go -source kobo-db.go
package pkg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/timchurchard/readstat/internal"
)

type KoboDatabase interface {
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

func (k koboDatabase) Contents() ([]KoboBook, error) {
	stmt, _, err := k.conn.Prepare(`SELECT ContentID, BookID, BookTitle, ReadStatus, WordCount, BookmarkWordOffset, ___PercentRead, TimeSpentReading, DateCreated FROM content`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// result := []KoboBook{}k

	for stmt.Step() {

		/*		if stmt.ColumnInt(4) < 1 { // Words
				continue
			}*/
		/*if stmt.ColumnInt(6) < 1 { // PercentRead
			continue
		}*/
		/*if stmt.ColumnInt(8) < 1 { // TimeSpent
			continue
		}*/

		fmt.Println("---------------------------------")
		fmt.Printf("%s / %s / %s\n", stmt.ColumnText(0), stmt.ColumnText(1), stmt.ColumnText(2)) // Contents ID, Book ID, Title
		fmt.Printf("ReadStatus %d / WordCount %d / BookmarkWords %d\n", stmt.ColumnInt(3), stmt.ColumnInt(4), stmt.ColumnInt(5))
		fmt.Printf("PercentRead %d\n", stmt.ColumnInt(6))
		fmt.Printf("TimeSpent %d\n", stmt.ColumnInt(7))
		fmt.Printf("DateCreated %s\n", stmt.ColumnText(8))

	}
	if err := stmt.Err(); err != nil {
		return nil, err
	}

	err = stmt.Close()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (k koboDatabase) Events() ([]KoboEvent, error) {
	stmt, _, err := k.conn.Prepare(`SELECT EventType, FirstOccurrence, LastOccurrence, EventCount, ContentID, ExtraData FROM Event`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// result := []KoboEvent{}

	for stmt.Step() {
		extraData := []byte{}
		colData := stmt.ColumnBlob(5, extraData)

		fmt.Println("---------------------------------")
		fmt.Printf("%d / %s / %s\n", stmt.ColumnInt(0), stmt.ColumnText(1), stmt.ColumnText(2))        // event type, first, last
		fmt.Printf("%d / %s / len(colData)=%d\n", stmt.ColumnInt(3), stmt.ColumnText(4), len(colData)) // count, content (fn), extradata blob

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

		json.NewEncoder(os.Stdout).Encode(v)
		if r.Len() != 0 {
			fmt.Println("not all read")
		}
	}
	if err := stmt.Err(); err != nil {
		return nil, err
	}

	err = stmt.Close()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (k koboDatabase) Close() error {
	return k.conn.Close()
}
