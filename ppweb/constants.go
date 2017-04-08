package main

import (
	"database/sql"

	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
)

// Header name
const HTTPCookieSession = "popcon_session"

// Location for Time
var Location = sctypes.Location

// Invalid String for sql.NullString
var NullStringInvalid = sql.NullString{Valid: false}

type ContextValueKeyType int

const (
	ContextValueKeySessionTemplateData ContextValueKeyType = iota
)
