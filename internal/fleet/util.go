package fleet

import (
	"fmt"
	"time"
)

func generateID(prefix string, seq uint64) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixMilli(), seq)
}

func now() time.Time {
	return time.Now().UTC()
}
