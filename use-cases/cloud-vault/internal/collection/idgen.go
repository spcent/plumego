package collection

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func newID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(rand.NewSource(time.Now().UnixNano()))).String()
}
