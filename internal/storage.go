//go:generate mockgen -package internal -destination storage_mock.go -source storage.go
package internal

import (
	"encoding/json"
	"os"
	"time"
)

type Storage struct {
	Devices  map[string]StorageDevice   `json:"devices"`
	Contents map[string]StorageContent  `json:"contents"`
	Events   map[string][]StorageEvents `json:"events"`

	fn string
}

type StorageDevice struct {
	Device string `json:"device"`
	Model  string `json:"model"`
}

type StorageContent struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	URL    string `json:"url"`

	Words int `json:"words"`

	IsBook     bool `json:"book"`
	IsFinished bool `json:"article_is_finished"`
}

type StorageEvents struct {
	EventName string `json:"event"`
	Time      string `json:"time"`
	Duration  int    `json:"duration"`
	Device    string `json:"device"`
}

const (
	StorageTimeFmt = "2006-01-02T15:05:06.000"
)

func OpenStorage(fn string) (Storage, error) {
	storage := Storage{
		Devices:  map[string]StorageDevice{},
		Contents: map[string]StorageContent{},
		Events:   map[string][]StorageEvents{},
		fn:       fn,
	}

	if _, err := os.Stat(fn); err == nil {
		storageBytes, err := os.ReadFile(fn)
		if err != nil {
			return storage, err
		}

		err = json.Unmarshal(storageBytes, &storage)
		if err != nil {
			return storage, err
		}
	}

	return storage, nil
}

func (s *Storage) Save() error {
	storageBytes, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(s.fn, storageBytes, 0o644)
}

func (s *Storage) AddContent(fn, title, author, url string, words int, book, finished bool) {
	s.Contents[fn] = StorageContent{
		ID:         fn,
		Title:      title,
		Author:     author,
		Words:      words,
		URL:        url,
		IsBook:     book,
		IsFinished: finished,
	}
}

func (s *Storage) AddDevice(device, model string) {
	s.Devices[device] = StorageDevice{
		Device: device,
		Model:  model,
	}
}

func (s *Storage) AddEvent(fn, device, name string, t time.Time, duration int) {
	timeStr := t.Format(StorageTimeFmt)

	found := false
	for eIdx := range s.Events[fn] {
		if s.Events[fn][eIdx].EventName == name && s.Events[fn][eIdx].Time == timeStr {
			found = true
		}
	}

	if !found {
		s.Events[fn] = append(s.Events[fn], StorageEvents{
			EventName: name,
			Time:      timeStr,
			Duration:  duration,
			Device:    device,
		})
	}
}
