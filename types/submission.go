package popconSCTypes

type SubmissionStatusType int

const (
	Accepted            SubmissionStatusType = 0
	WrongAnswer         SubmissionStatusType = 1
	TimeLimitExceeded   SubmissionStatusType = 2
	MemoryLimitExceeded SubmissionStatusType = 3
	RuntimeError        SubmissionStatusType = 4
	CompileError        SubmissionStatusType = 5
	InternalError       SubmissionStatusType = 6
	Judging             SubmissionStatusType = 7
	InQueue             SubmissionStatusType = 8
)

var StringToSubmissionStatusType = map[string]SubmissionStatusType{
	"Accepted":            Accepted,
	"WrongAnswer":         WrongAnswer,
	"TimeLimitExceeded":   TimeLimitExceeded,
	"MemoryLimitExceeded": MemoryLimitExceeded,
	"RuntimeError":        RuntimeError,
	"CompileError":        CompileError,
	"InternalError":       InternalError,
	"Judging":             Judging,
	"InQueue":             InQueue,
}

var SubmissionStatusTypeToString = map[SubmissionStatusType]string{
	Accepted:            "Accepted",
	WrongAnswer:         "WrongAnswer",
	TimeLimitExceeded:   "TimeLimitExceeded",
	MemoryLimitExceeded: "MemoryLimitExceeded",
	RuntimeError:        "RuntimeError",
	CompileError:        "CompileError",
	InternalError:       "InternalError",
	Judging:             "Judging",
	InQueue:             "InQueue",
}
