package uuid

import (
	crand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	mrand "math/rand"

	"github.com/google/uuid"
)

var maxInt64 = big.NewInt(math.MaxInt64)

// Next returns the next UUID using "github.com/google/uuid" library
// or the uuid 'degraded-cr-<rand int64>', or 'degraded-mr-<rand int64>'
// if the previous call errored.
func Next() string {
	uuid, err := uuid.NewRandom()
	if err != nil {
		r, err := crand.Int(crand.Reader, maxInt64)
		if err != nil {
			// Use math.rand.
			return fmt.Sprintf("degraded-mr-%d", mrand.Int63())
		}
		// Use crypto.rand.
		return fmt.Sprintf("degraded-cr-%d", r.Int64())
	}
	return uuid.String()
}
