package main

import (
	"math/rand"
	"time"

	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/seehuhn/mt19937"
)

type Judge interface {
	Run(ch chan<- JudgeStatus, tests <-chan TestCase, replaceNewlineChar bool)
}

type TestCase struct {
	ID    int64
	Input string
}

type ExecRequest struct {
	Image          string
	Cmd            []string
	SourceFileName string
	Env            []string
}

type JudgeStatus struct {
	Case   int64                        `json:"case"`
	Status sctypes.SubmissionStatusType `json:"jr"`
	Mem    int64                        `json:"mem"`
	Time   int64                        `json:"time"`
	Stdout string                       `json:"stdout"`
	Stderr string                       `json:"stderr"` // error and messageMsg
}

const TotalResultCaseID = -1

func CreateInternalError(id int64, msg string) JudgeStatus {
	return JudgeStatus{Case: id, Status: sctypes.SubmissionStatusInternalError, Stderr: msg}
}

const BASE_RAND_STRING = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomName() string {
	rng := rand.New(mt19937.New())
	rng.Seed(time.Now().UnixNano())

	res := make([]byte, 0, 16)
	for i := 0; i < 16; i++ {
		res = append(res, BASE_RAND_STRING[rng.Intn(len(BASE_RAND_STRING))])
	}

	return string(res)
}

func maxStatus(a, b sctypes.SubmissionStatusType) sctypes.SubmissionStatusType {
	_a := int(a)
	_b := int(b)
	if _a > _b {
		return a
	} else {
		return b
	}
}
