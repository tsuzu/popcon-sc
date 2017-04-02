package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
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
	Sid         int64            `gorm:"primary_key"`
	Pid         int64            `gorm:"not null;index"` //index
	Iid         int64            `gorm:"not null;index"` //index
	Lang        int64            `gorm:"not null"`
	Time        int64            `gorm:"not null"` //ms
	Mem         int64            `gorm:"not null"` //KB
	Score       int64            `gorm:"not null"`
	SubmitTime  time.Time        `gorm:"not null"`       //提出日時
	Status      SubmissionStatus `gorm:"not null;index"` //index
	MessageFile string           `gorm:"not null"`
	CodeFile    string           `gorm:"not null"`
	CasesFile   string           `gorm:"not null"`
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
	return dm.db.Where("pid=?", pid).Delete(Submission{}).Error
}

func (dm *DatabaseManager) SubmissionFind(sid int64) (*Submission, error) {
	var result Submission

	if err := dm.db.First(&result, sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUnknownSubmission
		}
		return nil, err
	}

	return &result, nil
}

func (dm *DatabaseManager) SubmissionUpdate(sid, time, mem int64, status SubmissionStatus, fin, all int, score int64) (ret error) {
	return dm.Begin(func(db *gorm.DB) error {
		var result Submission
		if err := db.First(&result, sid).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrUnknownSubmission
			}

			return err
		}

		result.Time = time
		result.Mem = mem
		result.Status = status
		result.Score = score

		return db.Save(&result).Error
	})

}

func (dm *DatabaseManager) SubmissionGetCode(sid int64) (string, error) {
	fp, err := os.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/code"), os.O_RDONLY, 0644)

	if err != nil {
		return "", err
	}

	defer fp.Close()

	b, err := ioutil.ReadAll(fp)

	if err != nil {
		return "", err
	}

	str := string(b)

	return str, nil
}

func (dm *DatabaseManager) SubmissionGetMsg(sid int64) (string, error) {
	var res string
	fm, err := FileManager.OpenFile(filepath.Join(SubmissionDir, strconv.FormatInt(sid, 10)+"/msg"), os.O_RDONLY, false)

	if err != nil {
		return "", ErrFileOpenFailed
	}

	defer fm.Close()

	b, err := ioutil.ReadAll(fm)

	if err != nil {
		return "", ErrFileOpenFailed
	}

	res = string(b)

	return res, nil
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

func (dm *DatabaseManager) SubmissionListWithPid(pid int64) (*[]Submission, error) {
	var results []Submission

	err := dm.db.Where("pid=?", pid).Find(&results).Error

	if err != nil {
		return nil, err
	}

	return &results, nil
}

type SubmissionView struct {
	SubmitTime    int64
	Cid           int64
	Pidx          int64
	Name          string
	Uid           string
	UserName      string
	Lang          string
	Score         int64
	RawStatus     SubmissionStatus
	Time          int64
	Mem           int64
	Sid           int64
	HighlightType string
	Iid           int64
	Status        string
}

// TODO: Gormに切り替え
func (dm *DatabaseManager) submissionViewQueryCreate(cid, iid, lid, pidx, stat int64, order string, offset, limit int64) (*gorm.DB, error) {
	db := dm.db.Model(&Submission{}).Joins("inner join contest_problems on submissions.pid = contest_problems.pid").Joins("inner join users on submissions.iid = users.iid").Joins("inner languages on submissions.lang=languages.lid")

	if cid != -1 {
		db = db.Where("contest_problems.cid=?", strconv.FormatInt(cid, 10))
	}

	if iid != -1 {
		db = db.Where("users.iid=?", strconv.FormatInt(iid, 10))
	}

	if pidx != -1 {
		if cid == -1 {
			return nil, ErrIllegalQuery
		}

		db = db.Where("contest_problems.pidx=?", strconv.FormatInt(pidx, 10))
	}

	if lid != -1 {
		db = db.Where("languages.lid=?", strconv.FormatInt(lid, 10))
	}

	if stat != -1 {
		db = db.Where("submissions.status=?", strconv.FormatInt(stat, 10))
	}

	if offset != -1 {
		db = db.Offset(offset)
	}
	if limit != -1 {
		db = db.Limit(limit)
	}

	if len(order) != 0 {
		db = db.Order(order)
	}

	return db, nil
}

func (dm *DatabaseManager) SubmissionViewCount(cid, iid, lid, pidx, stat int64) (int64, error) {
	//queryBase := "select count(submissions.sid) from submissions inner join contest_problems on submissions.pid = contest_problems.pid inner join users on submissions.iid = users.iid inner join languages on submissions.lang = languages.lid "
	db, err := dm.submissionViewQueryCreate(cid, iid, lid, pidx, stat, "", -1, -1)

	if err != nil {
		return 0, err
	}

	var cnt int64
	if err := db.Count(&cnt).Error; err != nil {
		return 0, err
	}

	return cnt, nil
}

func (dm *DatabaseManager) SubmissionViewList(cid, iid, lid, pidx, stat, offset, limit int64) ([]SubmissionView, error) {
	//queryBase := "select submissions.submit_time, contest_problems.cid, contest_problems.pidx, contest_problems.name, users.uid, users.user_name, languages.name, submissions.score, submissions.status, submissions.prog, submissions.time, submissions.mem, submissions.sid from submissions inner join contest_problems on submissions.pid = contest_problems.pid inner join user on submissions.iid = users.iid inner join languages on submissions.lang = languages.lid "
	db, err := dm.submissionViewQueryCreate(cid, iid, lid, pidx, stat, "submissions.sid desc", offset, limit)

	if err != nil {
		return nil, err
	}

	db = db.Select("submissions.submit_time, contest_problems.cid, contest_problems.pidx, contest_problems.name, users.uid, users.user_name, languages.name, submissions.score, submissions.status, submissions.time, submissions.mem, submissions.sid")

	var results []SubmissionView
	if err := db.Scan(&results).Error; err != nil {
		return nil, err
	}

	for i := range results {
		results[i].Status = results[i].RawStatus.String()

		if results[i].RawStatus == Judging {
			status, err := mainRM.JudgingProcessGet(results[i].Cid, results[i].Sid)

			if err != nil {
				DBLog.WithField("sid", results[i].Sid).WithField("cid", results[i].Cid).WithError(err).Error("JudgingProcessGet error")
			} else {
				results[i].Status = status
			}
		}
	}

	return results, nil
}

func (dm *DatabaseManager) SubmissionViewFind(sid, cid int64) (*SubmissionView, error) {
	//query := "select submissions.submit_time, contest_problems.cid, contest_problems.pidx, contest_problems.name, users.uid, users.user_name, languages.name, submissions.score, submissions.status, submissions.prog, submissions.time, submissions.mem, submissions.sid, languages.highlight_type, submissions.iid from submissions inner join contest_problems on submissions.pid = contest_problems.pid inner join user on submissions.iid = users.iid inner join languages on submissions.lang = languages.lid where submissions.sid = " + strconv.FormatInt(sid, 10)
	db, err := dm.submissionViewQueryCreate(cid, -1, -1, -1, -1, "", -1, 1)

	if err != nil {
		return nil, err
	}

	db = db.Select("submissions.submit_time, contest_problems.cid, contest_problems.pidx, contest_problems.name, users.uid, users.user_name, languages.name, submissions.score, submissions.status, submissions.time, submissions.mem, submissions.sid, languages.highlight_type, submissions.iid")

	var result SubmissionView
	if err := db.First(&result, sid).Error; err != nil {
		return nil, err
	}

	return &result, err
}
