package database

import (
	"strings"
)

func IsAlreadyExistsError(err error) bool {
	return strings.Contains(err.Error(), "already exists")
}

func IsDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "Duplicate")
}

type SqlScanIgnore struct{}

func (rc *SqlScanIgnore) Scan(v interface{}) error {
	return nil
}
