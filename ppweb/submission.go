package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cs3238-tsuzu/popcon-sc/ppweb/file_manager"
	"github.com/jinzhu/gorm"
)

// Database manager for Contest and Onlinejudge

//import "errors"

var SubmissionDir = "submissions/"

type SubmissionStatus int64

const (
	InQueue SubmissionStatus = iota
	Judging
	Accepted
	WrongAnswer
	TimeLimitExceeded
	MemoryLimitExceeded
	RuntimeError
	CompileError
	InternalError
)

func (ss SubmissionStatus) String() string {
	if v, ok := SubmissionStatusToString[ss]; ok {
		return v
	}

	return "<NA>"
}

var SubmissionStatusToString = map[SubmissionStatus]string{
	InQueue:             "WJ",
	Judging:             "JG",
	Accepted:            "AC",
	WrongAnswer:         "WA",
	TimeLimitExceeded:   "TLE",
	MemoryLimitExceeded: "MLE",
	RuntimeError:        "RE",
	CompileError:        "CE",
	InternalError:       "IE",
}

type Submission struct {
	Sid        int64            `gorm:"primary_key"`
	Pid        int64            `gorm:"not null;index"` //index
	Iid        int64            `gorm:"not null;index"` //index
	Lang       int64            `gorm:"not null"`
	Time       int64            `gorm:"not null"` //ms
	Mem        int64            `gorm:"not null"` //KB
	Score      int64            `gorm:"not null"`
	SubmitTime time.Time        `gorm:"not null"`       //提出日時
	Status     SubmissionStatus `gorm:"not null;index"` //index
	Prog       uint64           //テストケースの進捗状況(完了数<<32 & 全体数)
}

func (dm *DatabaseManager) CreateSubmissionTable() error {
	err := dm.db.AutoMigrate(&Submission{}).Error

	if err != nil {
		return err
	}

	return nil
}

func (dm *DatabaseManager) SubmissionAdd(pid, iid, lang int64, code string) (i int64, b error) {
	sm := Submission{
		Pid:        pid,
		Iid:        iid,
		Lang:       lang,
		SubmitTime: time.Now(),
		Status:     InQueue,
	}

	err := dm.db.Create(&sm).Error

	if err != nil {
		return 0, err
	}

	id := sm.Sid

	err = os.MkdirAll(filepath.Join(SubmissionDir, strconv.FormatInt(id, 10)), os.ModePerm)

	if err != nil {
		dm.SubmissionRemove(id)

		return 0, err
	}

	fp, err := os.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(id, 10)+"/msg"), os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		dm.SubmissionRemove(id)

		return 0, err
	}

	fp.Close()

	fp, err = os.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(id, 10)+"/.cases_lock"), os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		dm.SubmissionRemove(id)

		return 0, err
	}

	fp.Close()

	err = os.MkdirAll(filepath.Join(SubmissionDir, strconv.FormatInt(id, 10)+"/cases"), os.ModePerm)

	if err != nil {
		dm.SubmissionRemove(id)

		return 0, err
	}

	err = os.MkdirAll(filepath.Join(SubmissionDir, strconv.FormatInt(id, 10)), os.ModePerm)

	if err != nil {
		dm.SubmissionRemove(id)

		return 0, err
	}

	fp, err = os.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(id, 10)+"/code"), os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		dm.SubmissionRemove(id)

		return 0, err
	}

	defer fp.Close()

	_, err = fp.Write([]byte(code))

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (dm *DatabaseManager) SubmissionRemove(sid int64) error {
	sm := Submission{Sid: sid}
	err := dm.db.Delete(&sm).Error

	if err != nil {
		return err
	}

	fm1, _ := FileManager.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/.cases_lock"), os.O_RDONLY, true)
	fm2, _ := FileManager.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/msg"), os.O_RDONLY, true)

	defer func() {
		fm1.Close()
		fm2.Close()
	}()

	return os.RemoveAll(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)))
}

func (dm *DatabaseManager) SubmissionRemoveAll(pid int64) error {
	return dm.db.Delete(Submission{}, "pid=?", pid).Error
}

func (dm *DatabaseManager) SubmissionFind(sid int64) (*Submission, error) {
	var result Submission

	err := dm.db.First(&result, sid).Error

	if err == gorm.ErrRecordNotFound {
		return nil, ErrUnknownSubmission
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (dm *DatabaseManager) SubmissionUpdate(sid, time, mem int64, status SubmissionStatus, fin, all int, score int64) (ret error) {
	tx := dm.db.Begin()

	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if ret != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var result Submission
	err := tx.First(&result, sid).Error

	if err == gorm.ErrRecordNotFound {
		return ErrUnknownSubmission
	}

	if err != nil {
		return err
	}

	result.Time = time
	result.Mem = mem
	result.Status = status
	result.Prog = uint64(fin)<<32 | uint64(all)
	result.Score = score

	err = tx.Save(&result).Error

	return err
}

func (dm *DatabaseManager) SubmissionGetCode(sid int64) (*string, error) {
	fp, err := os.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/code"), os.O_RDONLY, 0644)

	if err != nil {
		return nil, err
	}

	defer fp.Close()

	b, err := ioutil.ReadAll(fp)

	if err != nil {
		return nil, err
	}

	str := string(b)

	return &str, nil
}

func (dm *DatabaseManager) SubmissionGetMsg(sid int64) *string {
	var res string
	fm, err := FileManager.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/msg"), os.O_RDONLY, false)

	if err != nil {
		return &res
	}

	defer fm.Close()

	b, err := ioutil.ReadAll(fm)

	if err != nil {
		return &res
	}

	res = string(b)

	return &res
}

func (dm *DatabaseManager) SubmissionSetMsg(sid int64, msg string) error {
	fm, err := FileManager.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/msg"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, true)

	if err != nil {
		return err
	}

	defer fm.Close()

	_, err = fm.Write([]byte(msg))

	return err
}

type SubmissionTestCase struct {
	Status SubmissionStatus
	Name   string
	Time   int64
	Mem    int64
}

func (dm *DatabaseManager) SubmissionGetCase(sid int64) (*map[int]SubmissionTestCase, error) {
	res := make(map[int]SubmissionTestCase)

	fm, err := FileManager.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/.cases_lock"), os.O_RDONLY, false)

	if err != nil {
		return nil, err
	}

	defer fm.Close()

	info, err := ioutil.ReadDir(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/cases"))

	if err != nil {
		return nil, err
	}

	for i := range info {
		if !info[i].IsDir() {
			id, err := strconv.ParseInt(info[i].Name(), 10, 64)

			if err != nil {
				continue
			}

			fp, err := os.Open(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/cases/"+info[i].Name()))

			if err != nil {
				continue
			}

			defer fp.Close()

			dec := json.NewDecoder(fp)

			var stc SubmissionTestCase

			err = dec.Decode(&stc)

			if err != nil {
				continue
			}

			res[int(id)] = stc
		}
	}

	return &res, nil
}

func (dm *DatabaseManager) SubmissionSetCase(sid int64, caseId int, stc SubmissionTestCase) error {
	fm, err := FileManager.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/.cases_lock"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, true)

	if err != nil {
		return err
	}

	defer fm.Close()

	fp, err := os.Create(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/cases/"+strconv.FormatInt(int64(caseId), 10)))

	if err != nil {
		return err
	}

	enc := json.NewEncoder(fp)

	return enc.Encode(stc)
}

func (dm *DatabaseManager) SubmissionClearCase(sid int64) error {
	fm, err := FileManager.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/.cases_lock"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, true)

	if err != nil {
		return err
	}

	defer fm.Close()

	err = os.RemoveAll(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/cases/"))

	if err != nil {
		return err
	}

	return os.MkdirAll(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/cases/"), os.ModePerm)
}

func (dm *DatabaseManager) SubmissionList(options ...[]interface{}) (*[]Submission, error) {
	var resulsts []Submission

	db := dm.db

	for i := range options {
		if len(options[i]) > 0 {
			db = db.Where(options[i][0], options[i][1:])
		}
	}

	err := db.Find(&resulsts).Error

	if err != nil {
		return nil, err
	}

	return &resulsts, nil
}

type SubmissionView struct {
	SubmitTime int64
	Cid        int64
	Pidx       int64
	Name       string
	Uid        string
	UserName   string
	Lang       string
	Score      int64
	Status     string
	Time       int64
	Mem        int64
	Sid        int64
}

// TODO: Gormに切り替え
func (dm *DatabaseManager) submissionViewQueryCreate(query string, cid, iid, lid, pidx, stat int64, order string, offset, limit int64) (string, error) {
	conditions := make([]string, 0, 5)

	if cid != -1 {
		conditions = append(conditions, "contest_problems.cid = "+strconv.FormatInt(cid, 10)+" ")
	}

	if iid != -1 {
		conditions = append(conditions, "users.iid = "+strconv.FormatInt(iid, 10)+" ")
	}

	if pidx != -1 {
		if cid == -1 {
			return "", errors.New("You must set cid to set pidx")
		}

		conditions = append(conditions, "contest_problems.pidx = "+strconv.FormatInt(pidx, 10)+" ")
	}

	if lid != -1 {
		conditions = append(conditions, "languages.lid = "+strconv.FormatInt(lid, 10)+" ")
	}

	if stat != -1 {
		conditions = append(conditions, "submissions.status = "+strconv.FormatInt(stat, 10)+" ")
	}

	where := strings.Join(conditions, "and ")

	if len(where) != 0 {
		where = "where " + where
	}

	var lim string
	if offset != -1 {
		lim = "limit " + strconv.FormatInt(offset, 10)
		if limit != -1 {
			lim = lim + ", " + strconv.FormatInt(limit, 10)
		}
	} else {
		if limit != -1 {
			lim = "limit " + strconv.FormatInt(limit, 10)
		}
	}

	query += where + order + lim

	return query, nil
}

func (dm *DatabaseManager) SubmissionViewCount(cid, iid, lid, pidx, stat int64) (int64, error) {
	queryBase := "select count(submissions.sid) from submissions inner join contest_problems on submissions.pid = contest_problems.pid inner join users on submissions.iid = users.iid inner join languages on submissions.lang = languages.lid "

	query, err := dm.submissionViewQueryCreate(queryBase, cid, iid, lid, pidx, stat, "", -1, -1)

	if err != nil {
		return 0, err
	}

	rows, err := dm.db.DB().Query(query)

	if err != nil {
		return 0, err
	}

	defer rows.Close()

	rows.Next()

	var cnt int64
	err = rows.Scan(&cnt)

	if err != nil {
		return 0, err
	}

	return cnt, err
}

func (dm *DatabaseManager) SubmissionViewList(cid, iid, lid, pidx, stat, offset, limit int64) (*[]SubmissionView, error) {
	queryBase := "select submissions.submit_time, contest_problems.cid, contest_problems.pidx, contest_problems.name, users.uid, users.user_name, languages.name, submissions.score, submissions.status, submissions.prog, submissions.time, submissions.mem, submissions.sid from submissions inner join contest_problems on submissions.pid = contest_problems.pid inner join user on submissions.iid = users.iid inner join languages on submissions.lang = languages.lid "

	query, err := dm.submissionViewQueryCreate(queryBase, cid, iid, lid, pidx, stat, "order by submissions.sid desc ", offset, limit)

	if err != nil {
		return nil, err
	}

	rows, err := dm.db.DB().Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	views := make([]SubmissionView, 0, 50)

	for rows.Next() {
		var sv SubmissionView

		var status int64
		var prog uint64
		rows.Scan(&sv.SubmitTime, &sv.Cid, &sv.Pidx, &sv.Name, &sv.Uid, &sv.UserName, &sv.Lang, &sv.Score, &status, &prog, &sv.Time, &sv.Mem, &sv.Sid)

		if status == int64(Judging) {
			all := prog & ((uint64(1) << 32) - 1)
			per := prog >> 32

			sv.Status = strconv.FormatInt(int64(per), 10) + "/" + strconv.FormatInt(int64(all), 10)
		} else {
			sv.Status = SubmissionStatusToString[SubmissionStatus(status)]
		}

		if status == int64(CompileError) || status == int64(InQueue) || status == int64(Judging) {
			sv.Mem = -1
			sv.Time = -1
			sv.Score = -1
		}

		views = append(views, sv)
	}

	return &views, nil
}

type SubmissionViewEach struct {
	SubmissionView
	HighlightType string
	Iid           int64
}

func (dm *DatabaseManager) SubmissionViewFind(sid int64) (*SubmissionViewEach, error) {
	query := "select submissions.submit_time, contest_problems.cid, contest_problems.pidx, contest_problems.name, users.uid, users.user_name, languages.name, submissions.score, submissions.status, submissions.prog, submissions.time, submissions.mem, submissions.sid, languages.highlight_type, submissions.iid from submissions inner join contest_problems on submissions.pid = contest_problems.pid inner join user on submissions.iid = users.iid inner join languages on submissions.lang = languages.lid where submissions.sid = " + strconv.FormatInt(sid, 10)

	rows, err := dm.db.DB().Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	rows.Next()

	var sv SubmissionViewEach
	var status int64
	var prog uint64

	err = rows.Scan(&sv.SubmitTime, &sv.Cid, &sv.Pidx, &sv.Name, &sv.Uid, &sv.UserName, &sv.Lang, &sv.Score, &status, &prog, &sv.Time, &sv.Mem, &sv.Sid, &sv.HighlightType, &sv.Iid)

	if err != nil {
		return nil, err
	}

	if status == int64(Judging) {
		all := prog & ((uint64(1) << 32) - 1)
		per := prog >> 32

		sv.Status = strconv.FormatInt(int64(per), 10) + "/" + strconv.FormatInt(int64(all), 10)
	} else {
		sv.Status = SubmissionStatusToString[SubmissionStatus(status)]
	}

	if status == int64(CompileError) || status == int64(InQueue) || status == int64(Judging) {
		sv.Mem = -1
		sv.Time = -1
		sv.Score = -1
	}

	return &sv, nil
}

func (dm *DatabaseManager) SubmissionGetPid(sid int64) (int64, error) {
	rows, err := dm.db.DB().Query("select pid from submissions where sid = ?", sid)

	if err != nil {
		return 0, err
	}

	defer rows.Close()

	var res int64
	rows.Next()
	err = rows.Scan(&res)

	return res, err
}
