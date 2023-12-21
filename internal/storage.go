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
	ID    string `json:"id"`
	Title string `json:"title"`

	Words int `json:"words"`

	URL    string `json:"url"`
	IsBook bool   `json:"book"`
}

type StorageEvents struct {
	EventName string `json:"event"`

	Time     string `json:"time"`
	Duration int    `json:"duration"`
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

func (s *Storage) AddContent(fn, title, url string, words int, book bool) {
	s.Contents[fn] = StorageContent{
		ID:     fn,
		Title:  title,
		Words:  words,
		URL:    url,
		IsBook: book,
	}
}

func (s *Storage) AddDevice(device, model string) {
	s.Devices[device] = StorageDevice{
		Device: device,
		Model:  model,
	}
}

func (s *Storage) AddEvent(fn, name string, t time.Time, duration int) {
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
		})
	}
}
