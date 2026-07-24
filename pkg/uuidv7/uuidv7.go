package uuidv7

import (
	"crypto/rand"
	"encoding/binary"
	"time"

	"github.com/google/uuid"
)

func New() uuid.UUID {
	return NewWithTime(time.Now())
}

func NewWithTime(t time.Time) uuid.UUID {
	var u uuid.UUID
	unixMs := uint64(t.UnixMilli())
	binary.BigEndian.PutUint32(u[0:4], uint32(unixMs>>16))
	binary.BigEndian.PutUint16(u[4:6], uint16(unixMs))
	_, _ = rand.Read(u[6:16])
	u[6] = (u[6] & 0x0f) | 0x70
	u[8] = (u[8] & 0x3f) | 0x80
	return u
}

func Parse(s string) (UUID, error) {
	return uuid.Parse(s)
}

func MustParse(s string) UUID {
	return uuid.MustParse(s)
}

type UUID = uuid.UUID

var Nil = uuid.Nil
