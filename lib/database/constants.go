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
var ErrUnknownRankingRow = errors.New("Unknown ranking row")
var ErrUnknownRankingCell = errors.New("Unknown ranking cell")
var ErrNotRegisteredRankingCell = errors.New("Not registered ranking cell")
var ErrUnknownTestcase = errors.New("Unknown testcase")
var ErrIllegalQuery = errors.New("Illegal Query")
var ErrKeyDuplication = errors.New("Key Duplication")
var ErrAlreadyTransactionBegun = errors.New("Transaction is already begun")
