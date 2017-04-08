package database

import "errors"

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
var ErrKeyDuplication = errors.New("Key Duplication")
