package sctypes

import (
	"database/sql"
	"encoding/json"
	"time"
)

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
	ContestTypeICPC
)

func (ct ContestType) String() string {
	return ContestTypeToString[ct]
}

var ContestTypeToString = map[ContestType]string{
	ContestTypeJOI:     "JOI",
	ContestTypePCK:     "PCK",
	ContestTypeAtCoder: "AtCoder",
	ContestTypeICPC:    "ICPC",
}

var accumulateAllScore = func(problems map[int64]RankingCell) int64 {
	var sum int64 = 0
	for i := range problems {
		sum += problems[i].Score
	}
	return sum
}

var ContestTypeCalculateGeneralScore = map[ContestType]func(problems map[int64]RankingCell) int64{
	ContestTypeJOI:     accumulateAllScore,
	ContestTypePCK:     accumulateAllScore,
	ContestTypeAtCoder: accumulateAllScore,
	ContestTypeICPC:    accumulateAllScore,
}

var maximizedTime = func(problems map[int64]RankingCell) time.Duration {
	var max time.Duration = 0
	for i := range problems {
		if problems[i].Time > max {
			max = problems[i].Time
		}
	}
	return max
}
var accumulateAllTime = func(problems map[int64]RankingCell) time.Duration {
	var sum time.Duration = 0
	for i := range problems {
		sum += problems[i].Time
	}
	return sum
}
var ContestTypeCalculateGeneralTime = map[ContestType]func(problems map[int64]RankingCell) time.Duration{
	ContestTypeJOI:     maximizedTime,
	ContestTypePCK:     maximizedTime,
	ContestTypeAtCoder: maximizedTime,
	ContestTypeICPC:    accumulateAllTime,
}

var accumulateAllPenalty = func(problems map[int64]RankingCell) int64 {
	var sum int64 = 0
	for i := range problems {
		sum += problems[i].Penalty
	}
	return sum
}
var ContestTypeCalculateGeneralPenalty = map[ContestType]func(problems map[int64]RankingCell) int64{
	ContestTypeJOI:     accumulateAllPenalty,
	ContestTypePCK:     accumulateAllPenalty,
	ContestTypeAtCoder: accumulateAllPenalty,
	ContestTypeICPC:    accumulateAllPenalty,
}
var ContestTypeToEvaluationFunction1 = map[ContestType]func(score, penalty, penaltySetting int64, time time.Duration) int64{
	ContestTypeJOI: func(score, penalty, penaltySetting int64, t time.Duration) int64 { return int64(-t) },
	ContestTypePCK: func(score, penalty, penaltySetting int64, t time.Duration) int64 { return -penalty },
	ContestTypeAtCoder: func(score, penalty, penaltySetting int64, t time.Duration) int64 {
		return -int64(t) - penalty*int64(time.Minute)
	},
	ContestTypeICPC: func(score, penalty, penaltySetting int64, t time.Duration) int64 {
		return -int64(t) - penalty*int64(time.Minute)
	},
}
var ContestTypeToEvaluationFunction2 = map[ContestType]func(score, penalty, penaltySetting int64, time time.Duration) int64{
	ContestTypeJOI:     func(score, penalty, penaltySetting int64, t time.Duration) int64 { return 0 },
	ContestTypePCK:     func(score, penalty, penaltySetting int64, t time.Duration) int64 { return int64(-t) },
	ContestTypeAtCoder: func(score, penalty, penaltySetting int64, t time.Duration) int64 { return 0 },
	ContestTypeICPC:    func(score, penalty, penaltySetting int64, t time.Duration) int64 { return 0 },
}

var ContestTypeFromString = map[string]ContestType{
	"JOI":     ContestTypeJOI,
	"PCK":     ContestTypePCK,
	"AtCoder": ContestTypeAtCoder,
	"ICPC":    ContestTypeICPC,
}

var ContestTypeCEPenalty = map[ContestType]bool{
	ContestTypeJOI:     false,
	ContestTypePCK:     false,
	ContestTypeAtCoder: false,
	ContestTypeICPC:    false,
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

type RankingCell struct {
	Valid          bool
	Sid, Jid       int64
	Time           time.Duration
	Score, Penalty int64
}

var InvalidRankingCell = RankingCell{Valid: false}

func (rc RankingCell) IsValid() bool {
	return rc.Valid
}

func (rc RankingCell) String() string {
	b, _ := json.Marshal(rc)

	return string(b)
}

func (rc *RankingCell) Parse(str string) {
	if len(str) == 0 {
		rc.Valid = false

		return
	}

	if err := json.Unmarshal([]byte(str), rc); err != nil {
		rc.Valid = false
	} else {
		rc.Valid = true
	}
}

func (rc *RankingCell) Scan(v interface{}) error {
	var str sql.NullString

	if err := str.Scan(v); err != nil {
		return err
	}

	rc.Parse(str.String)

	return nil
}

type RankingRow struct {
	Iid      int64
	Problems map[int64]RankingCell
	General  RankingCell
}

type RankingRowWithUserData struct {
	RankingRow
	Uid, UserName string
}
