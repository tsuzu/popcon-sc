package database

import (
	"strconv"
	"time"

	"fmt"

	"github.com/cs3238-tsuzu/popcon-sc/lib/filesystem"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/jinzhu/gorm"
	"gopkg.in/mgo.v2"
)

type SubmissionTestCase struct {
	ID     int64 `gorm:"primary_key"`
	Sid    int64 `gorm:"index"`
	Cid    int64 `gorm:"-"`
	Status sctypes.SubmissionStatusType
	CaseID int64
	Name   string
	Time   int64
	Mem    int64
}

func (stc SubmissionTestCase) TableName() string {
	return "submission_test_cases_" + strconv.FormatInt(stc.Cid, 10)
}

type Submission struct {
	Cid         int64                        `gorm:"-"`
	Sid         int64                        `gorm:"primary_key"`
	Pid         int64                        `gorm:"not null;index"` //index
	Iid         int64                        `gorm:"not null;index"` //index
	Lang        int64                        `gorm:"not null"`
	Time        int64                        `gorm:"not null"` //ms
	Mem         int64                        `gorm:"not null"` //KB
	Score       int64                        `gorm:"not null"`
	SubmitTime  time.Time                    `gorm:"not null;default:CURRENT_TIMESTAMP"` //提出日時
	Status      sctypes.SubmissionStatusType `gorm:"not null;index"`                     //index
	MessageFile string                       `gorm:"not null"`
	CodeFile    string                       `gorm:"not null"`
	Cases       []SubmissionTestCase         `gorm:"ForeignKey:Sid"`
}

func (s Submission) TableName() string {
	return "submissions_" + strconv.FormatInt(s.Cid, 10)
}

func (dm *DatabaseManager) CreateSubmissionTable() error {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		if v, ok := db.Get("gorm:association:source"); ok {
			if s, ok := v.(*Submission); ok {
				return "submission_test_cases_" + strconv.FormatInt(s.Cid, 10)
			} else if s, ok := v.(Submission); ok {
				return "submission_test_cases_" + strconv.FormatInt(s.Cid, 10)
			}
		}

		return defaultTableName
	}

	/*err := dm.db.AutoMigrate(&Submission{}, &SubmissionTestCase{}).Error

	if err != nil {
		return err
	}*/

	return nil
}

func (dm *DatabaseManager) SubmissionAutoMigrate(cid int64) error {
	return dm.db.AutoMigrate(&Submission{Cid: cid}, &SubmissionTestCase{Cid: cid}).Error
}

func (dm *DatabaseManager) SubmissionAdd(cid, pid, iid, lang int64, code string) (i int64, b error) {
	_, path, err := mainDB.fs.FileSecureUpdate(fs.FS_CATEGORY_SUBMISSION, "", code)

	sm := Submission{
		Cid:        cid,
		Pid:        pid,
		Iid:        iid,
		Lang:       lang,
		SubmitTime: time.Now(),
		Status:     sctypes.SubmissionStatusInQueue,
		CodeFile:   path,
	}

	err = dm.db.AutoMigrate(&sm, &SubmissionTestCase{Cid: cid}).Create(&sm).Error

	if err != nil {
		return 0, err
	}

	return sm.Sid, nil
}

func (dm *DatabaseManager) SubmissionRemove(cid, sid int64) error {
	return dm.Begin(func(db *gorm.DB) error {
		var result Submission
		result.Cid = cid

		if err := db.First(&result, sid).Error; err != nil {
			return err
		}

		if err := db.Model(&result).Association("Cases").Clear().Error; err != nil {
			return err
		}
		if err := dm.submissionTestCaseDeleteUnassociated(cid, db); err != nil {
			dm.Logger().WithError(err).Error("submissionTestCaseDeleteUnassociated() error")
		}

		if err := db.Delete(&result).Error; err != nil {
			return err
		}

		if err := mainDB.fs.RemoveLater(fs.FS_CATEGORY_SUBMISSION, result.CodeFile); err != nil {
			dm.fs.Logger().WithError(err).Error("RemoveLater() error")
		}

		if err := mainDB.fs.RemoveLater(fs.FS_CATEGORY_SUBMISSION_MSG, result.MessageFile); err != nil {
			dm.fs.Logger().WithError(err).Error("RemoveLater() error")
		}

		return nil
	})
}

func (dm *DatabaseManager) SubmissionRemoveAll(cid, pid int64) error {
	// TODO: implement
	return nil
	return dm.db.Delete(Submission{Cid: cid, Pid: pid}).Error
}

func (dm *DatabaseManager) SubmissionFind(cid, sid int64) (*Submission, error) {
	var result Submission
	result.Cid = cid

	if err := dm.db.First(&result, sid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUnknownSubmission
		}
		return nil, err
	}

	return &result, nil
}

func (dm *DatabaseManager) SubmissionUpdate(cid, sid, time, mem int64, status sctypes.SubmissionStatusType, fin, all int64, score int64) (ret error) {
	return dm.Begin(func(db *gorm.DB) error {
		var result Submission
		result.Cid = cid
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

		if status == sctypes.SubmissionStatusJudging {
			if err := dm.redis.JudgingProcessUpdate(sid, strconv.FormatInt(fin, 10)+"/"+strconv.FormatInt(all, 10)); err != nil {
				dm.Logger().WithError(err).Error("JudgingProcessUpdate() error")
			}
		} else if result.Status == sctypes.SubmissionStatusJudging {
			if err := dm.redis.JudgingProcessDelete(sid); err != nil {
				dm.Logger().WithError(err).Error("JudgingProcessDelete() error")
			}
		}

		return db.Save(&result).Error
	})
}

func (dm *DatabaseManager) SubmissionGetCode(cid, sid int64) (string, error) {
	var result Submission
	result.Cid = cid
	if err := dm.db.Select("code_file").First(&result, sid).Error; err != nil {
		return "", err
	}

	b, err := mainDB.fs.Read(fs.FS_CATEGORY_SUBMISSION, result.CodeFile)

	if err == mgo.ErrNotFound {
		return "", fs.ErrFileOpenFailed
	}

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (dm *DatabaseManager) SubmissionGetMsg(cid, sid int64) (string, error) {
	var result Submission
	result.Cid = cid
	if err := dm.db.Select("message_file").First(&result, sid).Error; err != nil {
		return "", err
	}

	if len(result.MessageFile) == 0 {
		return "", nil
	}

	b, err := mainDB.fs.Read(fs.FS_CATEGORY_SUBMISSION_MSG, result.MessageFile)

	if err == mgo.ErrNotFound {
		return "", fs.ErrFileOpenFailed
	}

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (dm *DatabaseManager) SubmissionSetMsg(cid, sid int64, msg string) error {
	return dm.Begin(func(db *gorm.DB) error {
		var result Submission
		result.Cid = cid
		if err := db.Select("message_file").First(&result, sid).Error; err != nil {
			return err
		}

		f, path, err := mainDB.fs.FileSecureUpdate(fs.FS_CATEGORY_SUBMISSION_MSG, result.MessageFile, msg)

		if err != nil {
			return err
		}

		if err := db.Model(&Submission{Sid: sid, Cid: cid}).Update("message_file", path).Error; err != nil {
			return err
		}

		f()
		return nil
	})
}

func (dm *DatabaseManager) SubmissionGetCase(cid, sid int64) ([]SubmissionTestCase, error) {
	var results []SubmissionTestCase
	if err := dm.db.Model(Submission{Sid: sid, Cid: cid}).Order("case_id asc").Related(&results, "Cases").Error; err != nil {
		return nil, err
	}

	return results, nil
}

func (dm *DatabaseManager) SubmissionSetCase(cid, sid int64, caseId int, stc SubmissionTestCase) error {
	if err := dm.db.Model(Submission{Sid: sid, Cid: cid}).Association("Cases").Append(stc).Error; err != nil {
		return err
	}

	return nil
}

func (dm *DatabaseManager) SubmissionClearCase(cid, sid int64) error {
	if err := dm.db.Model(Submission{Sid: sid, Cid: cid}).Association("Cases").Clear().Error; err != nil {
		return err
	}

	return dm.SubmissionTestCaseDeleteUnassociated(cid)
}

func (dm *DatabaseManager) submissionTestCaseDeleteUnassociated(cid int64, db *gorm.DB) error {
	return db.Where("sid IS NULL").Delete(SubmissionTestCase{Cid: cid}).Error
}

func (dm *DatabaseManager) SubmissionTestCaseDeleteUnassociated(cid int64) error {
	return dm.submissionTestCaseDeleteUnassociated(cid, dm.db)
}

func (dm *DatabaseManager) SubmissionListWithPid(cid, pid int64) ([]Submission, error) {
	var results []Submission

	err := dm.db.Model(Submission{Cid: cid}).Where("pid=?", pid).Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

type SubmissionView struct {
	SubmitTime    time.Time
	Cid           int64
	Pidx          int64
	Name          string
	Uid           string
	UserName      string
	Lang          string
	Score         int64
	RawStatus     sctypes.SubmissionStatusType
	Time          int64
	Mem           int64
	Sid           int64
	HighlightType string
	Iid           int64
	Status        string
}

// TODO: Gormに切り替え
func (dm *DatabaseManager) submissionViewQueryCreate(cid, iid, lid, pidx, stat int64, order string, offset, limit int64) (*gorm.DB, error) {
	table := Submission{Cid: cid}.TableName()
	db := dm.db.Table(table + " as submissions").Joins("inner join contest_problems on submissions.pid = contest_problems.pid").Joins("inner join users on submissions.iid = users.iid").Joins("inner join languages on submissions.lang=languages.lid")

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

	db = db.Select("submissions.submit_time, contest_problems.cid, contest_problems.pidx, contest_problems.name, users.uid, users.user_name, languages.name as lang, submissions.score, submissions.status, submissions.time, submissions.mem, submissions.sid")

	var results []SubmissionView
	if err := db.Scan(&results).Error; err != nil {
		return nil, err
	}

	for i := range results {
		results[i].Status = results[i].RawStatus.String()

		if results[i].RawStatus == sctypes.SubmissionStatusJudging {
			status, err := dm.redis.JudgingProcessGet(results[i].Sid)

			if err != nil {
				dm.Logger().WithField("sid", results[i].Sid).WithField("cid", results[i].Cid).WithError(err).Error("JudgingProcessGet error")
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

	db = db.Select("submissions.submit_time, contest_problems.cid, contest_problems.pidx, contest_problems.name, users.uid, users.user_name, languages.name as lang, submissions.score, submissions.status, submissions.time, submissions.mem, submissions.sid, languages.highlight_type, submissions.iid").Where("sid=?", sid)

	var result SubmissionView
	if err := db.First(&result).Error; err != nil {
		return nil, err
	}

	result.Status = result.RawStatus.String()
	fmt.Println(result)
	if result.RawStatus == sctypes.SubmissionStatusJudging {
		status, err := dm.redis.JudgingProcessGet(result.Sid)

		if err != nil {
			dm.Logger().WithField("sid", result.Sid).WithField("cid", result.Cid).WithError(err).Error("JudgingProcessGet error")
		} else {
			result.Status = status
		}
	}

	return &result, err
}
