package database

import (
	"time"

	"sync"

	"github.com/cs3238-tsuzu/popcon-sc/lib/filesystem"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/jinzhu/gorm"
)

type Contest struct {
	Cid             int64               `gorm:"primary_key"`
	Name            string              `gorm:"not null;unique_index"`
	StartTime       time.Time           `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	FinishTime      time.Time           `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	Admin           int64               `gorm:"not null"`
	Penalty         int64               `gorm:"not null"`
	Type            sctypes.ContestType `gorm:"not null"`
	DescriptionFile string              `gorm:"not null"`
	dm              *DatabaseManager    `gorm:"-"`
}

func (c *Contest) ProblemAdd(pidx int64, name string, time, mem int64, jtype sctypes.JudgeType) (*ContestProblem, error) {
	pb, err := c.dm.ContestProblemAdd(c.Cid, pidx, name, time, mem, jtype)

	if err != nil {
		return nil, err
	}

	return c.dm.ContestProblemFind(c.Cid, pb)
}

func (c *Contest) DescriptionUpdate(desc string) error {
	var res Contest
	return c.dm.Begin(func(db *gorm.DB) error {
		if err := db.Select("description_file").First(&res, c.Cid).Error; err != nil {
			return err
		}

		f, newName, err := c.dm.fs.FileSecureUpdate(fs.FS_CATEGORY_CONTEST_DESCRIPTION, res.DescriptionFile, desc)

		if err != nil {
			return err
		}

		res.Cid = c.Cid
		if err := db.Model(&res).Update("description_file", newName).Error; err != nil {
			return err
		}

		f()
		return nil
	})
}

func (c *Contest) DescriptionLoad() (string, error) {
	var res Contest
	res.Cid = c.Cid

	if err := c.dm.db.Select("description_file").First(&res, c.Cid).Error; err != nil {
		return "", err
	}

	b, err := c.dm.fs.Read(fs.FS_CATEGORY_CONTEST_DESCRIPTION, res.DescriptionFile)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (dm *DatabaseManager) CreateContestTable() error {
	err := dm.db.AutoMigrate(&Contest{}).Error

	if err != nil && !IsAlreadyExistsError(err) {
		return err
	}

	return nil
}

func (dm *DatabaseManager) ContestAdd(name string, start time.Time, finish time.Time, admin int64, ctype sctypes.ContestType, penalty int64, backgroundPreparation func(cid int64) error) (*Contest, error) {
	contest := Contest{
		Name:       name,
		StartTime:  start,
		FinishTime: finish,
		Admin:      admin,
		Type:       ctype,
		Penalty:    penalty,
		dm:         dm,
	}

	if err := dm.BeginDM(func(dm *DatabaseManager) error {
		err := dm.db.Create(&contest).Error

		if err != nil {
			return err
		}

		var bgError error
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			bgError = backgroundPreparation(contest.Cid)
		}()
		err = dm.SubmissionAutoMigrate(contest.Cid)

		if err != nil {
			return err
		}
		err = dm.ContestProblemAutoMigrate(contest.Cid)

		if err != nil {
			return err
		}

		wg.Wait()
		if bgError != nil {
			return bgError
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &contest, nil
}

func (dm *DatabaseManager) ContestUpdate(cid int64, name string, start time.Time, finish time.Time, admin int64, ctype sctypes.ContestType, penalty int64) error {
	cont := Contest{
		Cid:        cid,
		Name:       name,
		StartTime:  start,
		FinishTime: finish,
		Admin:      admin,
		Type:       ctype,
		Penalty:    penalty,
	}

	err := dm.db.Omit("description_file").Save(&cont).Error

	if err != nil {
		return err
	}

	return nil
}

func (dm *DatabaseManager) ContestDelete(cid int64) error {
	if cid == 0 {
		return ErrUnknownContest
	}

	var res Contest
	err := dm.db.Select("description_file").First(&res, cid).Error

	if err != nil {
		return err
	}

	err = dm.db.Delete(&Contest{Cid: cid}).Error

	if err != nil {
		return err
	}

	return dm.fs.Remove(fs.FS_CATEGORY_CONTEST_DESCRIPTION, res.DescriptionFile)
}

func (dm *DatabaseManager) ContestFind(cid int64) (*Contest, error) {
	var res Contest

	err := dm.db.First(&res, cid).Error

	if err == gorm.ErrRecordNotFound {
		return nil, ErrUnknownContest
	}

	if err != nil {
		return nil, err
	}
	res.dm = dm

	return &res, nil
}

func (dm *DatabaseManager) ContestGetType(cid int64) (sctypes.ContestType, error) {
	var res Contest

	err := dm.db.Select("type").First(&res, cid).Error

	if err == gorm.ErrRecordNotFound {
		return 0, ErrUnknownContest
	}

	if err != nil {
		return 0, err
	}

	return res.Type, nil
}

func (dm *DatabaseManager) ContestCount(options ...[]interface{}) (int64, error) {
	var count int64

	db := dm.db.Model(&Contest{})
	for i := range options {
		if len(options[i]) > 0 {
			db = db.Where(options[i][0], options[i][1:]...)
		}
	}

	err := db.Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

// ContestList : if "offset" and "limit" aren't neccesary, set -1
func (dm *DatabaseManager) ContestList(offset, limit int, options ...[]interface{}) (*[]Contest, error) {
	var results []Contest

	db := dm.db
	for i := range options {
		if len(options[i]) > 0 {
			db = db.Where(options[i][0], options[i][1:]...)
		}
	}

	err := db.Offset(offset).Limit(limit).Order("start_time asc").Find(&results).Error

	if err != nil {
		return nil, err
	}

	for i := range results {
		results[i].dm = dm
	}

	return &results, nil
}
