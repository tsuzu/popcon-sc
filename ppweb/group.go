package main

import "errors"

// Group is a struct to save GroupData
type Group struct {
	Gid  int64  `gorm:"primary_key"`
	Name string `gorm:"not null;unique_index"`
}

func (dm *DatabaseManager) CreateGroupTable() error {
	err := dm.db.AutoMigrate(&Group{}).Error

	if err != nil {
		return err
	}

	dm.GroupAdd("General")

	return nil
}

// GroupAdd adds a new group
// len(groupName) <= 50
func (dm *DatabaseManager) GroupAdd(name string) (int64, error) {
	if name == GroupAdminiStratorName {
		return 0, ErrKeyDuplication
	}

	if len(name) > 50 {
		return 0, errors.New("len(groupName) > 50")
	}

	group := Group{
		Name: name,
	}

	if err := dm.db.Create(&group).Error; err != nil {
		return 0, err
	}

	return group.Gid, nil
}

// GroupFind finds a group with groupID
func (dm *DatabaseManager) GroupFind(gid int64) (*Group, error) {
	if gid == 0 {
		return &Group{
			Gid:  0,
			Name: GroupAdminiStratorName,
		}, nil
	}

	var result Group

	if err := dm.db.Find(&result, gid).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

// GroupRemove removes from groups
func (dm *DatabaseManager) GroupRemove(gid int64) error {
	return dm.db.Delete(Group{}, gid).Error
}

func (dm *DatabaseManager) GroupList() ([]Group, error) {
	var results []Group

	if err := dm.db.Find(&results).Error; err != nil {
		return nil, err
	}

	res := make([]Group, len(results)+1)
	res[0] = Group{Gid: 0, Name: GroupAdminiStratorName}
	for i := range results {
		res[i+1] = results[i]
	}

	return res, nil
}
