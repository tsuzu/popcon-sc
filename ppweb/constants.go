package main

import (
	"database/sql"
	"errors"

	"github.com/cs3238-tsuzu/popcon-sc/types"
)

// Header name
const HTTPCookieSession = "popcon_session"

// Location for Time
var Location = sctypes.Location

// Invalid String for sql.NullString
var NullStringInvalid = sql.NullString{Valid: false}

// Error templates
var ErrUnknownProblem = errors.New("Unknown problem")
var ErrUnknownLanguage = errors.New("Unknown language")
var ErrUnknownSubmission = errors.New("Unknown submission")
var ErrUnknownContest = errors.New("Unknown contest")
var ErrUnknownGroup = errors.New("Unknown group")
var ErrUnknownUser = errors.New("Unknown user")
var ErrUnknownSession = errors.New("Unknown session")
var ErrUnknownTestcase = errors.New("Unknown testcase")
var ErrIllegalQuery = errors.New("Illegal Query")
var ErrFileOpenFailed = errors.New("Failed opening a file")
var ErrKeyDuplication = errors.New("Key Duplication")

type ContextValueKeyType int

const (
	ContextValueKeySessionTemplateData ContextValueKeyType = iota
)

const GroupAdministrator int64 = 0
const GroupAdminiStratorName = "Administrator"
