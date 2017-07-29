package sctypes

type SubmissionStatusType int

// Caution: Don't change this order without checking ../dataisase/submission.go
const (
	SubmissionStatusInQueue SubmissionStatusType = iota
	SubmissionStatusJudging
	SubmissionStatusAccepted
	SubmissionStatusWrongAnswer
	SubmissionStatusTimeLimitExceeded
	SubmissionStatusMemoryLimitExceeded
	SubmissionStatusRuntimeError
	SubmissionStatusCompileError
	SubmissionStatusInternalError
)

func (ss SubmissionStatusType) String() string {
	if v, ok := SubmissionStatusTypeToString[ss]; ok {
		return v
	}

	return "<NA>"
}

var SubmissionStatusTypeFromString = map[string]SubmissionStatusType{
	"WJ":  SubmissionStatusInQueue,
	"JG":  SubmissionStatusJudging,
	"AC":  SubmissionStatusAccepted,
	"WA":  SubmissionStatusWrongAnswer,
	"TLE": SubmissionStatusTimeLimitExceeded,
	"MLE": SubmissionStatusMemoryLimitExceeded,
	"RE":  SubmissionStatusRuntimeError,
	"CE":  SubmissionStatusCompileError,
	"IE":  SubmissionStatusInternalError,
}

var SubmissionStatusTypeToString = map[SubmissionStatusType]string{
	SubmissionStatusInQueue:             "WJ",
	SubmissionStatusJudging:             "JG",
	SubmissionStatusAccepted:            "AC",
	SubmissionStatusWrongAnswer:         "WA",
	SubmissionStatusTimeLimitExceeded:   "TLE",
	SubmissionStatusMemoryLimitExceeded: "MLE",
	SubmissionStatusRuntimeError:        "RE",
	SubmissionStatusCompileError:        "CE",
	SubmissionStatusInternalError:       "IE",
}
