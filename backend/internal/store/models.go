package store

import "time"

type Doc struct {
	ID        string
	Title     string
	Bytes     []byte
	Version   int64
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}
