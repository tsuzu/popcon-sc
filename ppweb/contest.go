package main

import (
	"time"

	"github.com/cs3238-tsuzu/popcon-sc/types"
	"github.com/jinzhu/gorm"
)

type Contest struct {
	Cid        int64                     `gorm:"primary_key"`
	Name       string                    `gorm:"not null;unique_index"`
	StartTime  time.Time                 `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	FinishTime time.Time                 `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	Admin      int64                     `gorm:"not null"`
	Type       popconSCTypes.ContestType `gorm:"not null"`
}

func (c *Contest) ProblemAdd(pidx int64, name string, time, mem int64, jtype JudgeType) (*ContestProblem, error) {
	pb, err := mainDB.ContestProblemAdd(c.Cid, pidx, name, time, mem, jtype)

	if err != nil {
		return nil, err
	}

	return mainDB.ContestProblemFind(pb)
}

func (dm *DatabaseManager) CreateContestTable() error {
	err := dm.db.AutoMigrate(&Contest{}).Error

	if err != nil {
		return err
	}

	return nil
}

func (dm *DatabaseManager) ContestAdd(name string, start time.Time, finish time.Time, admin int64, ctype popconSCTypes.ContestType) (int64, error) {
	contest := Contest{
		Name:       name,
		StartTime:  start,
		FinishTime: finish,
		Admin:      admin,
		Type:       ctype,
	}

	err := dm.db.Create(&contest).Error

	if err != nil {
		return 0, err
	}

	id := contest.Cid

	// TODO: ***Important***Support GridFS
	/*fm, err := FileManager.OpenFile(filepath.Join(ContestDir, strconv.FormatInt(id, 10)), os.O_CREATE|os.O_WRONLY, true)

	if err != nil {
		dm.ContestDelete(id)

		return 0, err
	}

	fm.Close()*/

	return id, nil
}

func (dm *DatabaseManager) ContestUpdate(cid int64, name string, start time.Time, finish time.Time, admin int64, ctype popconSCTypes.ContestType) error {
	cont := Contest{
		Cid:        cid,
		Name:       name,
		StartTime:  start,
		FinishTime: finish,
		Admin:      admin,
		Type:       ctype,
	}

	err := dm.db.Save(&cont).Error

	if err != nil {
		return err
	}

	return nil
}

func (dm *DatabaseManager) ContestDelete(cid int64) error {
	if cid == 0 {
		return ErrUnknownContest
	}

	err := dm.db.Delete(&Contest{Cid: cid}).Error

	if err != nil {
		return err
	}

	// TODO: Support GridFS
	/*fm, err := FileManager.OpenFile(filepath.Join(ContestDir, strconv.FormatInt(cid, 10)), os.O_WRONLY, true)

	if err != nil {
		return err
	}

	defer fm.Close()

	return os.Remove(filepath.Join(ContestDir, strconv.FormatInt(cid, 10)))*/

	return nil
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

	return &res, nil
}

func (dm *DatabaseManager) ContestDescriptionUpdate(cid int64, desc string) error {
	// TODO:Support GridFS

	/*fm, err := FileManager.OpenFile(filepath.Join(ContestDir, strconv.FormatInt(cid, 10)), os.O_WRONLY|os.O_TRUNC, true)

	if err != nil {
		return err
	}

	defer fm.Close()

	_, err = fm.Write([]byte(desc))

	return err*/
	return nil
}

func (dm *DatabaseManager) ContestDescriptionLoad(cid int64) (string, error) {
	//TODO: Support GridFS
	/*fm, err := FileManager.OpenFile(filepath.Join(ContestDir, strconv.FormatInt(cid, 10)), os.O_RDONLY, false)

	if err != nil {
		return "", err
	}

	defer fm.Close()

	b, err := ioutil.ReadAll(fm)

	if err != nil {
		return "", err
	}

	return string(b), err*/

	return "", nil
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
	var resulsts []Contest

	db := dm.db
	for i := range options {
		if len(options[i]) > 0 {
			db = db.Where(options[i][0], options[i][1:]...)
		}
	}

	err := db.Offset(offset).Limit(limit).Order("start_time asc").Find(&resulsts).Error

	if err != nil {
		return nil, err
	}

	return &resulsts, nil
}
