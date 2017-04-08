package database

import "errors"
import "time"

// "create table if not exists news (text varchar(256), unixTime int, index uti(unixTime))"
// News contains news showed on "/"
type News struct {
	Nid      int64     `gorm:"primary_key"`
	Text     string    `gorm:"size:4095"`
	UnixTime time.Time `gorm:"not null;index"`
}

func (dm *DatabaseManager) CreateNewsTable() error {
	err := dm.db.AutoMigrate(&News{}).Error

	if err != nil {
		return err
	}

	return nil
}

// NewsAdd adds a news displayed on "/"
// len(text) <= 256
func (dm *DatabaseManager) NewsAdd(text string) error {
	return dm.NewsAddWithTime(text, time.Now())
}

// NewsAddWithTime adds a news displayed on "/" with unixtime
// len(text) <= 256
func (dm *DatabaseManager) NewsAddWithTime(text string, unixTime time.Time) error {
	if len(text) > 4095 {
		return errors.New("len(text) > 4095")
	}

	news := News{
		Text:     text,
		UnixTime: time.Now(),
	}

	err := dm.db.Create(&news).Error

	return err
}

func (dm *DatabaseManager) NewsGet(cnt int) ([]News, error) {
	var resulsts []News

	err := dm.db.Order("unix_time desc").Limit(cnt).Find(&resulsts).Error

	if err != nil {
		return nil, err
	}

	return resulsts, nil
}
