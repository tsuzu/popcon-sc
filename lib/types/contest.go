package sctypes

import "time"

type ContestType int

const (
	ContestTypeJOI ContestType = iota
	// Score: Score(>)
	// Value1: Time(<)
	// Value2: None

	ContestTypePCK
	// Score: Score(>)
	// Value1: WA(<)
	// Value2: Time(<)

	ContestTypeAtCoder
	// Score: Score(>)
	// Value1: Time+Penalty(<)
	// Value2: None
)

func (ct ContestType) String() string {
	return ContestTypeToString[ct]
}

var ContestTypeToString = map[ContestType]string{
	ContestTypeJOI:     "JOI",
	ContestTypePCK:     "PCK",
	ContestTypeAtCoder: "AtCoder",
}

var ContestTypeToEvaluationFunction1 = map[ContestType]func(score, penalty int64, time time.Duration) int64{
	ContestTypeJOI:     func(score, penalty int64, time time.Duration) int64 { return 0 },
	ContestTypePCK:     func(score, penalty int64, time time.Duration) int64 { return 0 },
	ContestTypeAtCoder: func(score, penalty int64, time time.Duration) int64 { return 0 },
}
var ContestTypeToEvaluationFunction2 = map[ContestType]func(score, penalty int64, time time.Duration) int64{
	ContestTypeJOI:     func(score, penalty int64, time time.Duration) int64 { return 0 },
	ContestTypePCK:     func(score, penalty int64, time time.Duration) int64 { return 0 },
	ContestTypeAtCoder: func(score, penalty int64, time time.Duration) int64 { return 0 },
}

var ContestTypeFromString = map[string]ContestType{
	"JOI":     ContestTypeJOI,
	"PCK":     ContestTypePCK,
	"AtCoder": ContestTypeAtCoder,
}

var ContestTypeCEPenalty = map[ContestType]bool{
	ContestTypeJOI:     false,
	ContestTypePCK:     false,
	ContestTypeAtCoder: false,
}

type JudgeType int

const (
	JudgePerfectMatch JudgeType = iota
	JudgeRunningCode
	JudgeInteractive
)

var JudgeTypeToString = map[JudgeType]string{
	JudgePerfectMatch: "Perfect Match",
	JudgeRunningCode:  "Running Code",
	JudgeInteractive:  "Interactive",
}

var JudgeTypeFromString = map[string]JudgeType{
	"Perfect Match": JudgePerfectMatch,
	"Running Code":  JudgeRunningCode,
	"Interactive":   JudgeInteractive,
}
