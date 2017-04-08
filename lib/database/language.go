package database

import (
	"github.com/jinzhu/gorm"
)

type Language struct {
	Lid           int64  `gorm:"primary_key"`
	Name          string `gorm:"not null"`
	HighlightType string `gorm:"not null"`
	Active        bool   `gorm:"not null;default:true;index"`
}

func (dm *DatabaseManager) CreateLanguageTable() error {
	err := dm.db.AutoMigrate(&Language{}).Error

	if err != nil {
		return err
	}

	return nil
}

func (dm *DatabaseManager) LanguageAdd(name, highlightType string, active bool) (int64, error) {
	lang := Language{
		Name:          name,
		HighlightType: highlightType,
		Active:        active,
	}

	err := dm.db.Create(&lang).Error

	if err != nil {
		return 0, err
	}

	return lang.Lid, nil
}

func (dm *DatabaseManager) LanguageUpdate(lid int64, name, highlightType string, active bool) error {
	err := dm.db.Save(&Language{
		Lid:           lid,
		Name:          name,
		HighlightType: highlightType,
		Active:        active,
	}).Error

	return err
}

func (dm *DatabaseManager) LanguageFind(lid int64) (*Language, error) {
	var resulst Language

	err := dm.db.First(&resulst, lid).Error

	if err == gorm.ErrRecordNotFound {
		return nil, ErrUnknownLanguage
	}

	if err != nil {
		return nil, err
	}

	return &resulst, nil
}

func (dm *DatabaseManager) LanguageList() ([]Language, error) {
	var resulsts []Language

	err := dm.db.Find(&resulsts).Error

	if err != nil {
		return nil, err
	}

	return resulsts, nil
}

func (dm *DatabaseManager) LanguageActiveList() ([]Language, error) {
	var results []Language

	err := dm.db.Where("active=1").Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}
