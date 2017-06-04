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
