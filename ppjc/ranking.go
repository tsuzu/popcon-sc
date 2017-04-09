package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"time"

	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
)

var int64CompareFunction = func(a, b int64) int {
	switch {
	case a < b:
		return 1
	case a > b:
		return -1
	default:
		return 0
	}
}

var CompareFunctionForContest = []utils.Comparator{ // ContestType -> Comparation

	func(a, b interface{}) int { // JOI
		_a := a.(RankingScore)
		_b := b.(RankingScore)

		if _a.Score == _b.Score {
			return int64CompareFunction(_b.Score, _a.Score) // Reversed
		} else if _a.Value1 == _b.Value1 {
			return int64CompareFunction(_a.Value1, _b.Value1)
		}

		return int64CompareFunction(_a.Uid, _b.Uid)
	},
	func(a, b interface{}) int { // PCK
		_a := a.(RankingScore)
		_b := b.(RankingScore)

		if _a.Score == _b.Score {
			return int64CompareFunction(_b.Score, _a.Score) // Reversed
		} else if _a.Value1 == _b.Value1 {
			return int64CompareFunction(_a.Value1, _b.Value1)
		} else if _a.Value2 == _b.Value2 {
			return int64CompareFunction(_a.Value2, _b.Value2)
		}

		return int64CompareFunction(_a.Uid, _b.Uid)
	},
	func(a, b interface{}) int { // AtCoder
		_a := a.(RankingScore)
		_b := b.(RankingScore)

		if _a.Score == _b.Score {
			return int64CompareFunction(_b.Score, _a.Score) // Reversed
		} else if _a.Value1 == _b.Value1 {
			return int64CompareFunction(_a.Value1, _b.Value1)
		}

		return int64CompareFunction(_a.Uid, _b.Uid)
	},
	func(a, b interface{}) int { // ICPC
		_a := a.(RankingScore)
		_b := b.(RankingScore)

		if _a.Score == _b.Score {
			return int64CompareFunction(_b.Score, _a.Score) // Reversed
		} else if _a.Value1 == _b.Value1 {
			return int64CompareFunction(_a.Value1, _b.Value1)
		}

		return int64CompareFunction(_a.Uid, _b.Uid)
	},
}

var CalculateRankingScoreFunctionForContest = []func(map[int64]*ProblemResult, int64) RankingScore{ // ContestType -> Comparation

	func(problems map[int64]*ProblemResult, penalty int64) RankingScore { // JOI
		var res RankingScore
		for k := range problems {
			res.Score += problems[k].Score
			if res.Value1 < problems[k].Time {
				res.Value1 = problems[k].Time
			}
		}

		return res
	},
	func(problems map[int64]*ProblemResult, penalty int64) RankingScore { // PCK
		var res RankingScore
		for k := range problems {
			res.Score += problems[k].Score
			res.Value1 += problems[k].WA
			if res.Value2 < problems[k].Time {
				res.Value2 = problems[k].Time
			}
		}

		return res
	},
	func(problems map[int64]*ProblemResult, penalty int64) RankingScore { // AtCoder
		var res RankingScore
		for k := range problems {
			res.Score += problems[k].Score
			if res.Value1 < problems[k].Time {
				res.Value1 = problems[k].Time
			}
			res.Value2 += problems[k].WA
		}
		res.Value1 += res.Value2 * penalty
		res.Value2 = 0

		return res
	},
	func(problems map[int64]*ProblemResult, penalty int64) RankingScore { // ICPC
		var res RankingScore
		for k := range problems {
			res.Score += problems[k].Score
			res.Value1 += problems[k].Time + penalty*problems[k].WA
		}

		return res
	},
}

type RankingScore struct {
	Score  int64
	Value1 int64
	Value2 int64
	Uid    int64
}

type RankingScoreSaved []RankingScore

type RankingCell struct {
	Score       int64
	Time        int64
	WrongAnswer int
}

type UserResult struct {
	*RankingScore
	Problems map[int64]*ProblemResult
}

type ProblemResult struct {
	Score int64
	Time  int64
	WA    int64
	Sid   int64

	Submissions map[int64]*SubmissionResult
}

type SubmissionResult struct {
	Jid     int64
	Score   int64
	Status  sctypes.SubmissionStatusType
	Elapsed int64
}

type Ranking struct {
	Cid                            int64
	StartTime, FinishTime, Penalty int64
	ContestType                    sctypes.ContestType

	/* Key: RankingCell, Val: nil */
	Ranking *redblacktree.Tree    `json:"-"`
	Details map[int64]*UserResult `json:"-"`

	rankingInfoMutex sync.RWMutex
	rankingMutex     sync.RWMutex
}

func NewRanking(cid int64, startTime int64, finishTime, penalty int64, contestType sctypes.ContestType) (*Ranking, error) {
	r := Ranking{
		Cid:         cid,
		StartTime:   startTime,
		FinishTime:  finishTime,
		Penalty:     penalty,
		ContestType: contestType,
	}

	dir := filepath.Join(GeneralSetting.RankingSavedFolderPath, strconv.FormatInt(r.Cid, 10))
	err := os.MkdirAll(dir, 0770)

	if err != nil {
		return nil, err
	}

	b, _ := json.Marshal(r)

	err = ioutil.WriteFile(filepath.Join(dir, "setting.json"), b, 0770)

	if err != nil {
		return nil, err
	}

	return &r, nil
}

func LoadRanking(cid int64) (*Ranking, error) {
	dir := filepath.Join(GeneralSetting.RankingSavedFolderPath, strconv.FormatInt(cid, 10))

	var ranking Ranking
	if b, err := ioutil.ReadFile(filepath.Join(dir, "setting.json")); err != nil {
		return nil, err
	} else {
		if err := json.Unmarshal(b, &ranking); err != nil {
			return nil, err
		}
	}

	var saved RankingScoreSaved

	if b, err := ioutil.ReadFile(filepath.Join(dir, "ranking.json")); err == nil {
		json.Unmarshal(b, &saved)
	}

	ranking.Ranking = redblacktree.NewWith(CompareFunctionForContest[ranking.ContestType])

	if saved != nil {
		for idx := range saved {
			ranking.Ranking.Put(saved[idx], nil)
		}
	}

	if ranking.FinishTime > GetCurrentUnixTime() {
		ranking.loadDetails()
	}

	return &ranking, nil
}

func (r *Ranking) loadDetails() {
	dir := filepath.Join(GeneralSetting.RankingSavedFolderPath, strconv.FormatInt(r.Cid, 10))
	if b, err := ioutil.ReadFile(filepath.Join(dir, "details.json")); err == nil {
		if err := json.Unmarshal(b, &r.Details); err != nil {
			r.Details = make(map[int64]*UserResult)
		}
	} else {
		r.Details = make(map[int64]*UserResult)
	}
}

func (r *Ranking) save() {
	dir := filepath.Join(GeneralSetting.RankingSavedFolderPath, strconv.FormatInt(r.Cid, 10))
	b, _ := json.Marshal(r)

	if err := ioutil.WriteFile(filepath.Join(dir, "setting.json"), b, 0770); err != nil {
		RKLog().WithError(err).Error("WriteFile(setting.json) error")

		return
	}

	it := r.Ranking.Iterator()

	saved := make(RankingScoreSaved, 0, r.Ranking.Size())
	for it.Begin(); it.Next(); {
		saved = append(saved, it.Key().(RankingScore))
	}

	b, _ = json.Marshal(saved)

	if err := ioutil.WriteFile(filepath.Join(dir, "ranking.json"), b, 0770); err != nil {
		RKLog().WithError(err).Error("WriteFile(ranking.json error")

		return
	}

	if r.Details != nil {
		b, _ = json.Marshal(r.Details)

		if err := ioutil.WriteFile(filepath.Join(dir, "details.json"), b, 0770); err != nil {
			RKLog().WithError(err).Error("WriteFile(details.json) error")

			return
		}
	}
}

func (r *Ranking) Update(u models.ContestInfo) error {
	var start, finish time.Time
	var err error
	if start, err = time.ParseInLocation(time.RFC3339, *u.StartTime, Location); err != nil {
		return errors.New("startTime format error. " + err.Error())
	}
	if finish, err = time.ParseInLocation(time.RFC3339, *u.FinishTime, Location); err != nil {
		return errors.New("finishTime format error. " + err.Error())
	}

	if start.Unix() >= finish.Unix() {
		return errors.New("finishTime must be later than startTime")
	}

	r.rankingInfoMutex.Lock()
	if r.StartTime != start.Unix() {
		if start.Unix() <= GetCurrentUnixTime() {
			return errors.New("startTime must be later than now")
		} else {
			r.StartTime = start.Unix()
		}
	}
	if r.FinishTime != finish.Unix() {
		if finish.Unix() <= GetCurrentUnixTime() {
			return errors.New("finishTime must be later than now")
		} else {
			r.FinishTime = finish.Unix()
		}
	}
	r.rankingInfoMutex.Unlock()

	if r.StartTime > GetCurrentUnixTime() {
		r.rankingInfoMutex.Lock()
		r.ContestType = sctypes.ContestTypeFromString[*u.ContestType]
		r.Penalty = *u.Penalty
		r.rankingInfoMutex.Unlock()
	}
	return nil
}

func (r *Ranking) NewSubmissionResult(result models.SubmissionResult) error {
	submitTime, err := time.Parse(time.RFC3339, *result.SubmitTime)
	if err != nil {
		return errors.New("submitTime parse error. " + err.Error())
	}

	r.rankingMutex.Lock()
	defer r.rankingMutex.Unlock()

	if r.Details == nil {
		r.loadDetails()
	}

	if _, ok := r.Details[*result.UID]; !ok {
		return nil
	}

	if submitTime.Unix() >= r.FinishTime {
		return nil
	}

	isChanged := false
	if p, ok := r.Details[*result.UID].Problems[*result.Pid]; !ok {

		pr := &ProblemResult{
			*result.Score,
			submitTime.Unix() - r.StartTime,
			0,
			*result.Sid,
			make(map[int64]*SubmissionResult),
		}
		pr.Submissions[*result.Sid] = &SubmissionResult{*result.Jid, *result.Score, sctypes.SubmissionStatusTypeFromString[*result.Status], submitTime.Unix() - r.StartTime}
		r.Details[*result.UID].Problems[*result.Pid] = pr
		isChanged = true
	} else {
		if p.Score < *result.Score {
			if _, ok := p.Submissions[*result.Sid]; !ok {
				p.Sid = *result.Sid
				p.Time = submitTime.Unix() - r.StartTime // コンテストが開始しているためcontestInfoMutexは不要
				p.Score = *result.Score
				if p.Submissions == nil {
					p.Submissions = make(map[int64]*SubmissionResult)
				}
				p.Submissions[*result.Sid] = &SubmissionResult{*result.Jid, *result.Score, sctypes.SubmissionStatusTypeFromString[*result.Status], submitTime.Unix() - r.StartTime}

				var cnt int64 = 0
				for k := range p.Submissions {
					if k == *result.Sid {
						break
					}
					cnt++
				}
				p.WA = cnt

				isChanged = true
			} else {
				var cnt int64 = 0
				for k, _ := range p.Submissions {
					if k == *result.Sid {
						if p.Submissions[k].Jid < *result.Jid {
							p.Submissions[k] = &SubmissionResult{*result.Jid, *result.Score, sctypes.SubmissionStatusTypeFromString[*result.Status], submitTime.Unix() - r.StartTime}
							p.Score = *result.Score
							p.WA = cnt
							p.Sid = k
							p.Time = submitTime.Unix() - r.StartTime // コンテストが開始しているためcontestInfoMutexは不要
						}
						break
					}
					cnt++
				}
				isChanged = true
			}
		} else {
			if _, ok := p.Submissions[*result.Sid]; !ok {
				p.Submissions[*result.Sid] = &SubmissionResult{*result.Jid, *result.Score, sctypes.SubmissionStatusTypeFromString[*result.Status], submitTime.Unix() - r.StartTime} // コンテストが開始しているためcontestInfoMutexは不要

				var cnt int64 = 0
				for k := range p.Submissions {
					if k == p.Sid {
						break
					}
					cnt++
				}
				if p.WA != cnt {
					p.WA = cnt
					isChanged = true
				}
			} else {
				if p.Submissions[*result.Sid].Jid < *result.Jid {
					if *result.Sid != p.Sid {
						p.Submissions[*result.Sid].Score = *result.Score
						p.Submissions[*result.Sid].Status = sctypes.SubmissionStatusTypeFromString[*result.Status]
						p.Submissions[*result.Sid].Jid = *result.Jid
					} else {
						var cnt int64 = 0
						var scoreMax int64 = 0
						var sidMax int64 = 0
						var cntMax int64 = 0
						for k := range p.Submissions {
							if scoreMax < p.Submissions[k].Score {
								scoreMax = p.Submissions[k].Score
								sidMax = k
								cntMax = cnt
							}
							cnt++
						}
						p.WA = cntMax
						p.Score = scoreMax
						p.Sid = sidMax
						p.Time = p.Submissions[sidMax].Elapsed
						isChanged = true
					}
				}

			}
		}
	}

	if isChanged {
		score := CalculateRankingScoreFunctionForContest[r.ContestType](r.Details[*result.UID].Problems, r.Penalty)
		score.Uid = *result.UID

		oldScore := *r.Details[*result.UID].RankingScore
		r.Ranking.Remove(oldScore)

		r.Details[*result.UID].RankingScore = &score
		r.Ranking.Put(score, nil)
	}
	return nil
}

func (r *Ranking) runImpl() {

	if r.Ranking == nil {
		r.Ranking = redblacktree.NewWith(CompareFunctionForContest[r.ContestType])
	}

	updateRanking := time.NewTicker(time.Duration(GeneralSetting.UpdateRankingTerm) * time.Millisecond)
	terminateTimer := time.NewTicker(time.Duration(GeneralSetting.RankingRunningTerm) * time.Minute)
	savingTimer := time.NewTicker(time.Duration(GeneralSetting.SavingTerm) * time.Minute)
	defer updateRanking.Stop()
	defer terminateTimer.Stop()
	defer savingTimer.Stop()

	//	isChanged := false
	for {
		select {
		case <-updateRanking.C:

		case <-savingTimer.C:

		case <-terminateTimer.C:

		}
	}
}

// Start Goroutine for Ranking
func (r *Ranking) Run(
	update <-chan models.ContestInfo,
	submission <-chan models.SubmissionResult,
) {
	go r.runImpl()
}
