package goatquery

import (
	"fmt"
	"time"
)

func ParseDateTime(value string) (*time.Time, error) {
	val, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return &val, nil
	}

	val, err = time.Parse(time.DateOnly, value)
	if err == nil {
		return &val, nil
	}

	return nil, fmt.Errorf("could not parse datetime")
}
