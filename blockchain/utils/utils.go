package utils

import (
	"fmt"
	"time"
)

func GetCurrentTimestamp() string {
	ts := time.Now().Unix()
	return fmt.Sprintf("%d", ts)
}
