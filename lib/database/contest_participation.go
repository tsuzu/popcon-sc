package database

import (
	"github.com/jinzhu/gorm"
)

// TODO: 処理をppjqに移動するため不要となる

type ContestParticipation struct {
	Cpid  int64 `gorm:"primary_key"`
	Iid   int64 `gorm:"index"`
	Cid   int64 `gorm:"index"`
	Admin bool
}

func (dm *DatabaseManager) CreateContestParticipationTable() error {
	err := dm.db.AutoMigrate(&ContestParticipation{}).Error

	if err != nil && !IsAlreadyExistsError(err) {
		return err
	}

	dm.db.Model(ContestParticipation{}).AddUniqueIndex("unq", "iid", "cid")

	return nil
}

func (dm *DatabaseManager) ContestParticipationAdd(iid, cid int64) error {
	cp := ContestParticipation{
		Iid:   iid,
		Cid:   cid,
		Admin: false,
	}
	if err := dm.db.Create(&cp).Error; err != nil {
		if IsDuplicateError(err) {
			return nil
		}

		return err
	}

	return nil
}

func (dm *DatabaseManager) ContestParticipationAddAsAdmin(iid, cid int64) error {
	cp := ContestParticipation{
		Iid:   iid,
		Cid:   cid,
		Admin: true,
	}
	if err := dm.db.Create(&cp).Error; err != nil {
		if IsDuplicateError(err) {
			return nil
		}

		return err
	}

	return nil
}

func (dm *DatabaseManager) ContestParticipationDelete(iid, cid int64) error {
	if err := dm.db.Model(ContestParticipation{}).Where("iid=? and cid=?", iid, cid).Delete(ContestParticipation{}).Error; err != nil {
		return err
	}

	return nil
}

// ContestParticipationCheck returns (has joined, is admin, error)
func (dm *DatabaseManager) ContestParticipationCheck(iid, cid int64) (bool, bool, error) {
	var cp ContestParticipation

	if err := dm.db.Where("cid=? and iid=?", cid, iid).First(&cp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, false, nil
		}

		return false, false, err
	}

	return true, cp.Admin, nil
}

func (dm *DatabaseManager) ContestParticipationRemove(cid int64) error {
	return nil
}
