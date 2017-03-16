package main

import (
	"database/sql"
	"errors"

	"github.com/cs3238-tsuzu/popcon-sc/types"
)

const HTTPCookieSession = "popcon_session"

var Location = sctypes.Location

var NullStringInvalid = sql.NullString{Valid: false}

var ErrUnknownProblem = errors.New("Unknown problem")
var ErrUnknownLanguage = errors.New("Unknown language")
var ErrUnknownSubmission = errors.New("Unknown submission")
var ErrUnknownContest = errors.New("Unknown contest")
var ErrUnknownGroup = errors.New("Unknown group")
var ErrUnknownUser = errors.New("Unknown user")
var ErrUnknownSession = errors.New("Unknown session")
var ErrUnknownTestcase = errors.New("Unknown testcase")
