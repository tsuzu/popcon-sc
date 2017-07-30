package database

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	mgo "gopkg.in/mgo.v2"

	"io"

	"strings"

	"github.com/cs3238-tsuzu/popcon-sc/lib/filesystem"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/cs3238-tsuzu/popcon-sc/lib/utility"
	"github.com/jinzhu/gorm"
)

type ContestProblemTestCase struct {
	Id     int64 `gorm:"primary_key"`
	Pid    int64
	Name   string `gorm:"size:128"`
	Input  string `gorm:"size:128"`
	Output string `gorm:"size:128"`
	Cid    int64  `gorm:"-"`
}

func (cptc ContestProblemTestCase) TableName() string {
	return "contest_problem_test_cases_" + strconv.FormatInt(cptc.Cid, 10)
}

type ContestProblemScoreSetCasesString string

func (cs *ContestProblemScoreSetCasesString) Get() []int64 {
	arr := strings.Split(string(*cs), ",")
	ret := make([]int64, len(arr))

	for i := range arr {
		ret[i], _ = strconv.ParseInt(arr[i], 10, 64)
	}

	return ret
}

func (cs *ContestProblemScoreSetCasesString) Set(a []int64) {
	arr := make([]string, len(a))

	for i := range a {
		arr[i] = strconv.FormatInt(a[i], 10)
	}

	*cs = ContestProblemScoreSetCasesString(strings.Join(arr, ","))
}

type ContestProblemScoreSet struct {
	Id             int64 `gorm:"primary_key"`
	Pid            int64
	CasesRawString string `gorm:"size:6143"`
	Score          int64
	Cases          ContestProblemScoreSetCasesString `gorm:"-"`
	Cid            int64                             `gorm:"-"`
}

func (cpss ContestProblemScoreSet) TableName() string {
	return "contest_problem_score_sets_" + strconv.FormatInt(cpss.Cid, 10)
}

func (ss *ContestProblemScoreSet) BeforeSave() error {
	ss.CasesRawString = string(ss.Cases)
	return nil
}

func (ss *ContestProblemScoreSet) AfterFind() error {
	ss.Cases = ContestProblemScoreSetCasesString(ss.CasesRawString)
	return nil
}

type ContestProblem struct {
	Pid           int64                    `gorm:"primary_key"`
	Cid           int64                    `gorm:"-"` //`gorm:"not null;index;unique_index:cid_and_pidx_index"`
	Pidx          int64                    `gorm:"not null;index;unique_index:cid_and_pidx_index"`
	Name          string                   `gorm:"not null;size:255"`
	Time          int64                    `gorm:"not null"` // Second
	Mem           int64                    `gorm:"not null"` // MB
	LastModified  int64                    `gorm:"not null"`
	Score         int64                    `gorm:"not null"`
	Type          sctypes.JudgeType        `gorm:"not null"`
	StatementFile string                   `gorm:"not null;size:128"`
	CheckerFile   string                   `gorm:"not null;size:128"`
	Cases         []ContestProblemTestCase `gorm:"ForeignKey:Pid"`
	Scores        []ContestProblemScoreSet `gorm:"ForeignKey:Pid"`
	dm            *DatabaseManager         `gorm:"-"`
}

func (dm *DatabaseManager) setCidForContestProblems(cid int64, arr []ContestProblem) {
	for i := range arr {
		arr[i].Cid = cid
		arr[i].dm = dm
	}
}

func (cp ContestProblem) TableName() string {
	return "contest_problems_" + strconv.FormatInt(cp.Cid, 10)
}

func (dm *DatabaseManager) CreateContestProblemTable() error {
	/*err := dm.db.AutoMigrate(&ContestProblem{}, &ContestProblemTestCase{}, &ContestProblemScoreSet{}).Error
	if err != nil {
		return err
	}*/

	prevHandler := gorm.DefaultTableNameHandler

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		if v, ok := db.Get("gorm:association:source"); ok {
			if reflect.TypeOf(db.Value) == reflect.TypeOf(&[]ContestProblemTestCase{}) {
				return "contest_problem_test_cases_" + strconv.FormatInt(v.(*ContestProblem).Cid, 10)
			}
			if reflect.TypeOf(db.Value) == reflect.TypeOf(&[]ContestProblemScoreSet{}) {
				return "contest_problem_score_sets_" + strconv.FormatInt(v.(*ContestProblem).Cid, 10)
			}
		}

		return prevHandler(db, defaultTableName)
	}

	return nil
}

func (dm *DatabaseManager) ContestProblemAutoMigrate(cid int64) error {
	return dm.db.AutoMigrate(ContestProblem{Cid: cid}, ContestProblemTestCase{Cid: cid}, ContestProblemScoreSet{Cid: cid}).Error
}

func (cp *ContestProblem) UpdateStatement(text string) error {
	return cp.dm.Begin(func(db *gorm.DB) error {
		res := *cp
		err := db.Select("statement_file").First(&res, cp.Pid).Error

		if err != nil {
			return err
		}

		suc, path, err := cp.dm.fs.FileSecureUpdate(fs.FS_CATEGORY_PROBLEM_STATEMENT, res.StatementFile, text)

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
	res := *cp
	err := cp.dm.db.Select("statement_file").First(&res, cp.Pid).Error

	if err != nil {
		return "", err
	}

	b, err := cp.dm.fs.Read(fs.FS_CATEGORY_PROBLEM_STATEMENT, res.StatementFile)

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

	return cp.dm.Begin(func(db *gorm.DB) error {
		res := *cp
		err := db.Select("checker_file").First(&res, cp.Pid).Error

		if err != nil {
			return err
		}

		suc, path, err := cp.dm.fs.FileSecureUpdate(fs.FS_CATEGORY_PROBLEM_CHECKER, res.CheckerFile, string(b))

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
	ret = cp.dm.Begin(func(db *gorm.DB) error {
		res := *cp
		err := db.Select("checker_file").First(&res, cp.Pid).Error

		if err != nil {
			return err
		}

		b, err := cp.dm.fs.Read(fs.FS_CATEGORY_PROBLEM_CHECKER, res.CheckerFile)

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

	defer cp.dm.ClearUnassociatedData(cp.Cid)
	return cp.dm.BeginDM(func(dm *DatabaseManager) error {
		f := func() {}
		var cases []ContestProblemTestCase
		var scores []ContestProblemScoreSet

		if err := dm.db.Model(cp).Related(&cases, "Cases").Related(&scores, "Scores").Error; err != nil {
			return err
		}

		if len(cases) != len(newCaseNames) {
			if len(cases) > len(newCaseNames) {
				oldFiles := make([]string, 0, (len(cases)-len(newCaseNames))*2)
				for i := len(newCaseNames); i < len(cases); i++ {
					oldFiles = append(oldFiles, cases[i].Input)
					oldFiles = append(oldFiles, cases[i].Output)
				}

				f = utility.FunctionJoin(f, func() {
					go func() {
						for i := range oldFiles {
							if err := cp.dm.fs.RemoveLater(fs.FS_CATEGORY_TESTCASE_INOUT, oldFiles[i]); err != nil {
								cp.dm.Logger().WithField("category", fs.FS_CATEGORY_TESTCASE_INOUT).WithField("path", oldFiles[i]).WithError(err).Error("Failed removing a file")
							}
						}
					}()
				})
			}
		}
		newCases := make([]ContestProblemTestCase, len(newCaseNames))

		for i := 0; i < len(newCaseNames); i++ {
			if i < len(cases) {
				newCases[i].Input = cases[i].Input
				newCases[i].Output = cases[i].Output
				newCases[i].Id = cases[i].Id
			}
			newCases[i].Name = newCaseNames[i]
			newCases[i].Pid = cp.Pid
			newCases[i].Cid = cp.Cid
		}

		for i := 0; i < len(newScores); i++ {
			if i < len(scores) {
				newScores[i].Id = scores[i].Id
				newScores[i].Pid = scores[i].Pid
			}
			newScores[i].Cid = cp.Cid
		}

		if err := dm.db.Model(cp).Association("Cases").Replace(&newCases).Error; err != nil {
			return err
		}
		if err := dm.db.Model(cp).Association("Scores").Replace(&newScores).Error; err != nil {
			return err
		}

		dm.db.Model(cp).Update("score", scoreSum)

		f()
		return nil
	})
}

func (cp *ContestProblem) CreateUniquelyNamedFile() (*mgo.GridFile, error) {
	id, err := cp.dm.redis.UniqueFileID(fs.FS_CATEGORY_TESTCASE_INOUT)

	if err != nil {
		return nil, err
	}

	return cp.dm.fs.Open(fs.FS_CATEGORY_TESTCASE_INOUT, "testcase_"+strconv.FormatInt(id, 10))
}

// ErrUnknownTestcase
func (cp *ContestProblem) UpdateTestCase(isInput bool, caseID int64, reader io.Reader) error {
	return cp.dm.Begin(func(db *gorm.DB) error {
		var cpcase ContestProblemTestCase
		cpcase.Cid = cp.Cid
		cpcase.Pid = cp.Pid
		if err := db.Model(cp).Offset(caseID).Limit(1).Order("id asc").Related(&cpcase, "Cases").Error; err != nil {
			return err
		}

		var f func()
		var p string
		var err error
		if isInput {
			f, p, err = cp.dm.fs.FileSecureUpdateWithReader(fs.FS_CATEGORY_TESTCASE_INOUT, cpcase.Input, reader)
			cpcase.Input = p
		} else {
			f, p, err = cp.dm.fs.FileSecureUpdateWithReader(fs.FS_CATEGORY_TESTCASE_INOUT, cpcase.Output, reader)
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
	var cpcase []ContestProblemTestCase

	if err := cp.dm.db.Model(cp).Offset(caseID).Limit(1).Order("id asc").Related(&cpcase, "Cases").Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUnknownTestcase
		}

		return nil, err
	}

	if len(cpcase) == 0 {
		return nil, ErrUnknownTestcase
	}

	var path string
	if isInput {
		path = cpcase[0].Input
	} else {
		path = cpcase[0].Output
	}
	if len(path) == 0 {
		return &utility.FakeEmptyReadCloser{}, nil
	}

	fp, err := cp.dm.fs.OpenOnly(fs.FS_CATEGORY_TESTCASE_INOUT, path)

	if err == mgo.ErrNotFound {
		return nil, ErrUnknownTestcase
	}

	return fp, err
}

func (cp *ContestProblem) LoadTestCases() ([]ContestProblemTestCase, []ContestProblemScoreSet, error) {
	var scores []ContestProblemScoreSet
	var cases []ContestProblemTestCase

	return cases, scores, cp.dm.BeginIfNotStarted(func(dm *gorm.DB) error {
		return dm.Model(cp).Related(&cases, "Cases").Related(&scores, "Scores").Error
	})
}

func (cp *ContestProblem) LoadTestCaseInfo(caseID int64) (int64, int64, error) {
	var cpcase []ContestProblemTestCase

	if err := cp.dm.db.Model(cp).Offset(caseID).Limit(1).Order("id asc").Related(&cpcase, "Cases").Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, 0, ErrUnknownTestcase
		}

		return 0, 0, err
	}
	if len(cpcase) == 0 {
		return 0, 0, ErrUnknownTestcase
	}

	var in, out int64 = 0, 0
	if len(cpcase[0].Input) != 0 {
		fp, err := cp.dm.fs.OpenOnly(fs.FS_CATEGORY_TESTCASE_INOUT, cpcase[0].Input)

		if err == nil {
			in = fp.Size()
		}
		fp.Close()
	}

	if len(cpcase[0].Output) != 0 {
		fp, err := cp.dm.fs.OpenOnly(fs.FS_CATEGORY_TESTCASE_INOUT, cpcase[0].Output)

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

	if err := cp.dm.db.Model(cp).Related(&cases, "Cases").Related(&scores, "Scores").Error; err != nil {
		return nil, nil, err
	}

	caseNames := make([]string, len(cases))

	for i := range cases {
		caseNames[i] = cases[i].Name
	}

	return caseNames, scores, nil
}

func (dm *DatabaseManager) ContestProblemAdd(cid, pidx int64, name string, timeLimit, mem int64, jtype sctypes.JudgeType) (int64, error) {
	cp := &ContestProblem{
		Cid:  cid,
		Pidx: pidx,
		Name: name,
		Time: timeLimit,
		Mem:  mem,
		Type: jtype,
	}

	err := dm.db.Create(cp).Error

	if err != nil {
		return 0, err
	}
	cp.dm = dm

	return cp.Pid, nil
}

func (dm *DatabaseManager) ContestProblemUpdate(prob ContestProblem) error {
	return dm.db.Save(&prob).Error
}

func (dm *DatabaseManager) ContestProblemDelete(cid, pid int64) error {
	log := dm.Logger().WithField("cid", cid).WithField("pid", pid)

	cp, err := dm.ContestProblemFind(cid, pid)

	if err != nil {
		return err
	}

	var cases []ContestProblemTestCase
	if err := dm.db.Model(cp).Related(&cases, "Cases").Error; err != nil {
		return err
	}

	for i := range cases {
		if err := cp.dm.fs.RemoveLater(fs.FS_CATEGORY_TESTCASE_INOUT, cases[i].Input); err != nil {
			log.WithError(err).WithField(fs.FS_CATEGORY_TESTCASE_INOUT, cases[i].Input).Error("RemoveLater() error")
		}
		if err := cp.dm.fs.RemoveLater(fs.FS_CATEGORY_TESTCASE_INOUT, cases[i].Output); err != nil {
			log.WithError(err).WithField(fs.FS_CATEGORY_TESTCASE_INOUT, cases[i].Output).Error("RemoveLater() error")
		}
	}

	model := dm.db.Model(cp)

	if err := model.Association("Cases").Clear().Error; err != nil {
		log.WithError(err).Error("Delete associations of cases")
	}
	if err := model.Association("Scores").Clear().Error; err != nil {
		log.WithError(err).Error("Delete associations of scores")
	}

	if err := dm.ClearUnassociatedData(cid); err != nil {
		log.WithError(err).Error("Failed Deleting unassociated data")
	}

	return dm.db.Delete(cp).Error
}

func (dm *DatabaseManager) ClearUnassociatedData(cid int64) error {
	return dm.db.Where("pid IS NULL").Delete(ContestProblemTestCase{Cid: cid}).Delete(ContestProblemScoreSet{Cid: cid}).Error
}

func (dm *DatabaseManager) ContestProblemFind(cid, pid int64) (*ContestProblem, error) {
	var res ContestProblem
	res.Cid = cid

	err := dm.db.First(&res, pid).Error

	if err == gorm.ErrRecordNotFound {
		return nil, ErrUnknownProblem
	}

	if err != nil {
		return nil, err
	}
	res.dm = dm

	return &res, nil
}

func (dm *DatabaseManager) ContestProblemFind2(cid, pidx int64) (*ContestProblem, error) {
	var res ContestProblem
	res.Cid = cid

	err := dm.db.Model(ContestProblem{Cid: cid}).Where("pidx=?", pidx).First(&res).Error

	if err == gorm.ErrRecordNotFound {
		return nil, ErrUnknownProblem
	}
	if err != nil {
		return nil, err
	}
	res.dm = dm

	return &res, nil
}

func (dm *DatabaseManager) ContestProblemList(cid int64) ([]ContestProblem, error) {
	var results []ContestProblem

	err := dm.db.Table(ContestProblem{Cid: cid}.TableName()).Order("pidx asc").Find(&results).Error

	if err != nil {
		return nil, err
	}

	dm.setCidForContestProblems(cid, results)

	return results, nil
}

func (dm *DatabaseManager) ContestProblemCount(cid int64) (int64, error) {
	var count int64

	err := dm.db.Table(ContestProblem{Cid: cid}.TableName()).Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (dm *DatabaseManager) ContestProblemListLight(cid int64) ([]ContestProblem, error) {
	var results []ContestProblem

	err := dm.db.Table(ContestProblem{Cid: cid}.TableName()).Select("pidx, name").Find(&results).Error

	if err != nil {
		return nil, err
	}

	// set cid
	dm.setCidForContestProblems(cid, results)

	return results, nil
}

func (dm *DatabaseManager) ContestProblemDeleteAll(cid int64) error {
	var results []ContestProblem

	if err := dm.db.Table(ContestProblem{Cid: cid}.TableName()).Select("pid").Find(&results).Error; err != nil {
		return err
	}

	for i := range results {
		dm.ContestProblemDelete(cid, results[i].Pid)
	}
	return nil
}

func (dm *DatabaseManager) ContestProblemRemoveAllWithTable(cid int64) error {
	query := "SELECT statement_file, checker_file FROM " + ContestProblem{Cid: cid}.TableName()

	rows, err := dm.db.CommonDB().Query(query)

	if err != nil {
		return err
	}

	statementFiles, checkerFiles := make([]string, 0, 50), make([]string, 0, 50)
	for rows.Next() {
		var statement, checker string
		rows.Scan(&statement, &checker)

		if len(statement) != 0 {
			statementFiles = append(statementFiles, statement)
		}
		if len(checker) == 0 {
			checkerFiles = append(checkerFiles, checker)
		}
	}
	rows.Close()

	if _, err := dm.db.CommonDB().Exec(fmt.Sprintf("DROP TABLE %s, %s, %s", ContestProblem{Cid: cid}.TableName(), ContestProblemTestCase{Cid: cid}.TableName(), ContestProblemScoreSet{Cid: cid}.TableName())); err != nil {
		return err
	}

	for i := range statementFiles {
		if err := dm.fs.RemoveLater(fs.FS_CATEGORY_PROBLEM_STATEMENT, statementFiles[i]); err != nil {
			dm.logger().WithError(err).Error("RemoveLater() error")
		}
	}
	for i := range checkerFiles {
		if err := dm.fs.RemoveLater(fs.FS_CATEGORY_PROBLEM_CHECKER, checkerFiles[i]); err != nil {
			dm.logger().WithError(err).Error("RemoveLater() error")
		}
	}

	return nil
}
