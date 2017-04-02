package main

import (
	"encoding/json"
	"strconv"

	mgo "gopkg.in/mgo.v2"

	"io"

	"github.com/jinzhu/gorm"
)

type JudgeType int

const (
	JudgePerfectMatch JudgeType = iota
	JudgeRunningCode
)

type ContestProblemTestCase struct {
	Id     int64 `gorm:"parimary_key"`
	Pid    int64
	Name   string `gorm:"size:128"`
	Input  string `gorm:"size:128"`
	Output string `gorm:"size:128"`
}

type ContestProblemScoreSetCasesString string

func (cs *ContestProblemScoreSetCasesString) Get() []int64 {
	var results []int64
	json.Unmarshal([]byte(*cs), &results)

	return results
}

func (cs *ContestProblemScoreSetCasesString) Set(a []int64) {
	if b, err := json.Marshal(a); err != nil {
		*cs = "[]"
	} else {
		*cs = ContestProblemScoreSetCasesString(string(b))
	}
}

type ContestProblemScoreSet struct {
	Id    int64 `gorm:"parimary_key"`
	Pid   int64
	Cases ContestProblemScoreSetCasesString `gorm:"size:6143"`
	Score int64
}

type ContestProblem struct {
	Pid           int64                    `gorm:"primary_key"`
	Cid           int64                    `gorm:"not null;index;unique_index:cid_and_pidx_index"`
	Pidx          int64                    `gorm:"not null;index;unique_index:cid_and_pidx_index"`
	Name          string                   `gorm:"not null;size:255"`
	Time          int64                    `gorm:"not null"` // Second
	Mem           int64                    `gorm:"not null"` // MB
	LastModified  int64                    `gorm:"not null"`
	Score         int                      `gorm:"not null"`
	Type          JudgeType                `gorm:"not null"` // int->JudgeType
	StatementFile string                   `gorm:"not null;size:128"`
	CheckerFile   string                   `gorm:"not null;size:128"`
	Cases         []ContestProblemTestCase `gorm:"ForeignKey:Pid"`
	Scores        []ContestProblemScoreSet `gorm:"ForeignKey:Pid"`
}

// TODO: テストケースの情報を乗っけるようにする途中でORMの変更が入ったので中断
// 適当にSQLに乗っけるように変更

func (cp *ContestProblem) UpdateStatement(text string) error {
	return mainDB.Begin(func(db *gorm.DB) error {
		var res ContestProblem
		err := db.Select("statement_file").First(&res, cp.Pid).Error

		if err != nil {
			return err
		}

		suc, path, err := mainFS.FileSecureUpdate(FS_CATEGORY_PROBLEM_STATEMENT, res.StatementFile, text)

		if err != nil {
			return err
		}

		err = db.Model(cp).Update("statement_file", path).Error

		if err != nil {
			return err
		}

		suc() // Finish successfully, then call this

		return nil
	})
}

func (cp *ContestProblem) LoadStatement() (string, error) {
	var res ContestProblem
	err := mainDB.db.Select("statement_file").First(&res, cp.Pid).Error

	if err != nil {
		return "", err
	}

	b, err := mainFS.Read(FS_CATEGORY_PROBLEM_STATEMENT, res.StatementFile)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

type CheckerSavedFormat struct {
	Lid  int64
	Code string
}

func (cp *ContestProblem) UpdateChecker(lid int64, code string) error {
	b, err := json.Marshal(CheckerSavedFormat{lid, code})

	if err != nil {
		return err
	}

	return mainDB.Begin(func(db *gorm.DB) error {
		var res ContestProblem
		err := db.Select("checker_file").First(&res, cp.Pid).Error

		if err != nil {
			return err
		}

		suc, path, err := mainFS.FileSecureUpdate(FS_CATEGORY_PROBLEM_CHECKER, res.CheckerFile, string(b))

		if err != nil {
			return err
		}

		err = db.Model(cp).Update("checker_file", path).Error

		if err != nil {
			return err
		}

		suc() // Finish successfully, then call this

		return nil
	})
}

func (cp *ContestProblem) LoadChecker() (lid int64, code string, ret error) {
	ret = mainDB.Begin(func(db *gorm.DB) error {
		var res ContestProblem
		err := db.Select("checker_file").First(&res, cp.Pid).Error

		if err != nil {
			return err
		}

		b, err := mainFS.Read(FS_CATEGORY_PROBLEM_CHECKER, res.CheckerFile)

		if err != nil {
			return err
		}

		var csf CheckerSavedFormat
		err = json.Unmarshal(b, &csf)

		if err != nil {
			return err
		}
		lid = csf.Lid
		code = csf.Code

		return nil
	})

	return
}

func (cp *ContestProblem) UpdateTestCaseNames(newCaseNames []string, newScores []ContestProblemScoreSet) error {
	var scoreSum int64 = 0
	for i := range newScores {
		scoreSum += newScores[i].Score
	}

	return mainDB.Begin(func(db *gorm.DB) error {
		f := func() {}
		var cases []ContestProblemTestCase
		var scores []ContestProblemScoreSet

		if err := db.Model(&cp).Related(&cases, "Cases").Related(&scores, "Scores").Error; err != nil {
			return err
		}

		if len(cases) != len(newCaseNames) {
			if len(cases) > len(newCaseNames) {
				oldFiles := make([]string, 0, (len(cases)-len(newCaseNames))*2)
				for i := len(newCaseNames); i < len(cases); i++ {
					oldFiles = append(oldFiles, cases[i].Input)
					oldFiles = append(oldFiles, cases[i].Output)
				}

				f = FunctionJoin(f, func() {
					for i := range oldFiles {
						if err := mainFS.Remove(FS_CATEGORY_TESTCASE_INOUT, oldFiles[i]); err != nil {
							FSLog.WithField("category", FS_CATEGORY_TESTCASE_INOUT).WithField("path", oldFiles[i]).WithError(err).Error("Failed removing a file")
						}
					}
				})
			}
		}
		newCases := make([]ContestProblemTestCase, len(newCaseNames))

		for i := 0; i < len(newCaseNames) && i < len(cases); i++ {
			if i < len(cases) {
				newCases[i].Input = cases[i].Input
				newCases[i].Output = cases[i].Output
				newCases[i].Id = cases[i].Id
			}
			newCases[i].Name = newCaseNames[i]
			newCases[i].Pid = cp.Pid
		}
		cases = newCases

		for i := 0; i < len(scores); i++ {
			newScores[i].Id = scores[i].Id
		}

		if err := db.Model(&cp).Association("Cases").Replace(cases).Replace(scores).Error; err != nil {
			return err
		}

		return nil
	})
}

func (cp *ContestProblem) CreateUniquelyNamedFile() (*mgo.GridFile, error) {
	id, err := mainRM.UniqueFileID(FS_CATEGORY_TESTCASE_INOUT)

	if err != nil {
		return nil, err
	}

	return mainFS.Open(FS_CATEGORY_TESTCASE_INOUT, "testcase_"+mainFS.TestcaseFileBaseTag+"_"+strconv.FormatInt(id, 10))
}

// ErrUnknownTestcase
func (cp *ContestProblem) UpdateTestCase(isInput bool, caseID int64, reader io.Reader) error {
	return mainDB.Begin(func(db *gorm.DB) error {
		var cpcase ContestProblemTestCase
		if err := db.Model(&cp).Offset(caseID).Limit(1).Order("id asc").Related(&cpcase, "Cases").Error; err != nil {
			return err
		}

		var f func()
		var p string
		var err error
		if isInput {
			f, p, err = mainFS.FileSecureUpdateWithReader(FS_CATEGORY_TESTCASE_INOUT, cpcase.Input, reader)
			cpcase.Input = p
		} else {
			f, p, err = mainFS.FileSecureUpdateWithReader(FS_CATEGORY_TESTCASE_INOUT, cpcase.Output, reader)
			cpcase.Output = p
		}

		if err != nil {
			return err
		}

		db.Save(&cpcase)

		f()
		return nil
	})

}

func (cp *ContestProblem) LoadTestCase(isInput bool, caseID int) (io.ReadCloser, error) {
	var cpcase ContestProblemTestCase

	if err := mainDB.db.Model(&cp).Offset(caseID).Limit(1).Order("id asc").Related(&cpcase, "Cases").Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUnknownTestcase
		}

		return nil, err
	}

	var path string
	if isInput {
		path = cpcase.Input
	} else {
		path = cpcase.Output
	}

	return mainFS.OpenOnly(FS_CATEGORY_TESTCASE_INOUT, path)
}

func (cp *ContestProblem) LoadTestCases() ([]ContestProblemTestCase, []ContestProblemScoreSet, error) {
	var scores []ContestProblemScoreSet
	var cases []ContestProblemTestCase

	if err := mainDB.db.Model(&cp).Related(&cases, "Cases").Related(&scores, "Scores").Error; err != nil {
		return nil, nil, err
	}

	return cases, scores, nil
}

func (cp *ContestProblem) LoadTestCaseInfo(caseID int) (int64, int64, error) {
	var cpcase ContestProblemTestCase

	if err := mainDB.db.Model(&cp).Offset(caseID).Limit(1).Order("id asc").Related(&cpcase, "Cases").Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, 0, ErrUnknownTestcase
		}

		return 0, 0, err
	}

	var in, out int64 = 0, 0
	if len(cpcase.Input) != 0 {
		fp, err := mainFS.OpenOnly(FS_CATEGORY_TESTCASE_INOUT, cpcase.Input)

		if err == nil {
			in = fp.Size()
		}
		fp.Close()
	}

	if len(cpcase.Output) != 0 {
		fp, err := mainFS.OpenOnly(FS_CATEGORY_TESTCASE_INOUT, cpcase.Output)

		if err == nil {
			out = fp.Size()
		}
		fp.Close()
	}

	return in, out, nil
}

func (cp *ContestProblem) LoadTestCaseNames() ([]string, []ContestProblemScoreSet, error) {
	var scores []ContestProblemScoreSet
	var cases []ContestProblemTestCase

	if err := mainDB.db.Model(&cp).Related(&cases, "Cases").Related(&scores, "Scores").Error; err != nil {
		return nil, nil, err
	}

	caseNames := make([]string, len(cases))

	for i := range cases {
		caseNames[i] = cases[i].Name
	}

	return caseNames, scores, nil
}

func (dm *DatabaseManager) CreateContestProblemTable() error {
	err := dm.db.AutoMigrate(&ContestProblem{}, &ContestProblemTestCase{}, &ContestProblemScoreSet{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (dm *DatabaseManager) ContestProblemAdd(cid, pidx int64, name string, timeLimit, mem int64, jtype JudgeType) (int64, error) {
	cp := ContestProblem{
		Cid:  cid,
		Pidx: pidx,
		Name: name,
		Time: timeLimit,
		Mem:  mem,
		Type: jtype,
	}

	err := mainDB.db.Create(&cp).Error

	if err != nil {
		return 0, err
	}

	return cp.Pid, nil
}

func (dm *DatabaseManager) ContestProblemUpdate(prob ContestProblem) error {
	return dm.db.Save(&prob).Error
}

func (dm *DatabaseManager) ContestProblemDelete(pid int64) error {
	log := DBLog.WithField("pid", pid)

	cp, err := dm.ContestProblemFind(pid)

	if err != nil {
		return err
	}

	var cases []ContestProblemTestCase
	if err := dm.db.Model(&cp).Related(cases, "Cases").Error; err != nil {
		return err
	}

	for i := range cases {
		mainFS.RemoveLater(FS_CATEGORY_TESTCASE_INOUT, cases[i].Input)
		mainFS.RemoveLater(FS_CATEGORY_TESTCASE_INOUT, cases[i].Output)
	}

	model := dm.db.Model(&cp)

	if err := model.Association("Cases").Clear().Error; err != nil {
		log.WithError(err).Error("Delete associations of cases")
	}
	if err := model.Association("Scores").Clear().Error; err != nil {
		log.WithError(err).Error("Delete associations of scores")
	}

	if err := dm.ClearUnassociatedData(); err != nil {
		log.WithError(err).Error("Failed Deleting unassociated data")
	}

	return dm.db.Delete(&cp).Error
}

func (dm *DatabaseManager) ClearUnassociatedData() error {
	return dm.db.Where("pid IS NULL").Delete(ContestProblemTestCase{}).Delete(ContestProblemScoreSet{}).Error
}

func (dm *DatabaseManager) ContestProblemFind(pid int64) (*ContestProblem, error) {
	var res ContestProblem

	err := dm.db.First(&res, pid).Error

	if err == gorm.ErrRecordNotFound {
		return nil, ErrUnknownProblem
	}

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (dm *DatabaseManager) ContestProblemFind2(cid, pidx int64) (*ContestProblem, error) {
	var res ContestProblem

	err := dm.db.Where("pidx=?", pidx).Where("cid=?", cid).First(&res).Error

	if err == gorm.ErrRecordNotFound {
		return nil, ErrUnknownProblem
	}
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (dm *DatabaseManager) ContestProblemList(cid int64) ([]ContestProblem, error) {
	var results []ContestProblem

	err := dm.db.Where("cid=?", cid).Order("pidx asc").Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (dm *DatabaseManager) ContestProblemCount(cid int64) (int64, error) {
	var count int64

	err := dm.db.Model(&ContestProblem{}).Where("cid=?", cid).Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (dm *DatabaseManager) ContestProblemListLight(cid int64) ([]ContestProblem, error) {
	var results []ContestProblem

	err := dm.db.Select("pidx", "name").Where("cid=?", cid).Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (dm *DatabaseManager) ContestProblemDeleteAll(cid int64) error {
	var results []ContestProblem

	if err := dm.db.Select("pid").Where("cid=?", cid).Find(&results).Error; err != nil {
		return err
	}

	for i := range results {
		dm.ContestProblemDelete(results[i].Pid)
	}
	return nil
}
