package storage

import (
	"time"
)

type Service interface {
	Save(string, time.Time) (string, error)
	Load(string) (string, error)
	LoadInfo(string) (*Item, error)
	IsAvailable(id uint64) bool
	Close() error
}

type Item struct {
	Id         uint64 `json:"id" redis:"id"`
	URL        string `json:"url" redis:"url"`
	Expiration string `json:"expires" redis:"expires"`
	Visits     uint64 `json:"visits" redis:"visits"`
}
