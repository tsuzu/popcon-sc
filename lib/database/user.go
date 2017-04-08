package database

import (
	"crypto/sha512"
	"database/sql"
	"errors"
	"strconv"

	"github.com/cs3238-tsuzu/popcon-sc/lib/utility"
	"github.com/jinzhu/gorm"
)

// User is a struct to save UserData
type User struct {
	Iid      int64          `gorm:"primary_key"`
	Uid      string         `gorm:"not null;size:128;unique_index"`
	UserName string         `gorm:"not null;size:256;unique_index"`
	PassHash []byte         `gorm:"not null"`
	Email    sql.NullString `gorm:"size:128;unique_index"`
	Gid      int64
	Enabled  bool `gorm:"not null;default:true"`
}

func (dm *DatabaseManager) CreateUserTable() error {
	err := dm.db.AutoMigrate(&User{}).Error

	if err != nil {
		return err
	}

	return err
}

// UserAdd adds a new user
func (dm *DatabaseManager) UserAdd(uid string, userName string, pass string, email sql.NullString, gid int64, enabled bool) (int64, error) {
	if len(uid) > 128 {
		return 0, errors.New("error: len(uid) > 128")
	}

	if len(userName) > 256 {
		return 0, errors.New("error: len(userName) > 256")
	}

	if len(pass) > 100 {
		return 0, errors.New("error: len(pass) > 100")
	}
	passHashArr := sha512.Sum512([]byte(pass))
	//passHash := hex.EncodeToString(sha512.Sum512([]byte(pass)))

	if email.Valid && len(email.String) > 128 {
		return 0, errors.New("error: len(email) > 128")
	}

	user := User{
		Uid:      uid,
		UserName: userName,
		PassHash: passHashArr[:],
		Email:    email,
		Gid:      gid,
		Enabled:  enabled,
	}

	err := dm.db.Create(&user).Error

	if err != nil {
		return 0, err
	}

	return user.Iid, nil
}

// UserUpdate updates the user with iid
func (dm *DatabaseManager) UserUpdate(iid int64, uid string, userName string, pass string, email string, groupID int64, enabled bool) error {
	if len(uid) > 128 {
		return errors.New("error: len(uid) > 128")
	}

	if len(userName) > 256 {
		return errors.New("error: len(userName) > 256")
	}

	if len(pass) > 100 {
		return errors.New("error: len(pass) > 100")
	}
	passHashArr := sha512.Sum512([]byte(pass))

	if len(email) > 128 {
		return errors.New("error: len(email) > 128")
	}

	err := dm.db.Save(&User{
		Uid:      uid,
		UserName: userName,
		PassHash: passHashArr[:],
		Email:    utility.NullStringCreate(email),
		Gid:      groupID,
		Enabled:  enabled,
	}).Error

	return err
}

func (dm *DatabaseManager) UserUpdateEnabled(iid int64, enabled bool) error {
	err := dm.db.Model(&User{Iid: iid}).Update("enabled", enabled).Error

	return err
}

func (dm *DatabaseManager) UserUpdatePassword(iid int64, pass string) error {
	passHashArr := sha512.Sum512([]byte(pass))
	err := dm.db.Model(&User{Iid: iid}).Update("pass_hash", passHashArr[:]).Error

	return err
}

// UserFind returns a User object
func (dm *DatabaseManager) userFind(key string, value string) (*User, error) {
	var result User

	err := dm.db.Where(key+"=?", value).First(&result).Error

	if err == gorm.ErrRecordNotFound {
		return nil, ErrUnknownUser
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UserFindFromIID finds an User with iid
func (dm *DatabaseManager) UserFindFromIID(iid int64) (*User, error) {
	return dm.userFind("iid", strconv.FormatInt(iid, 10))
}

// UserFindFromUserID finds an User with Uid
func (dm *DatabaseManager) UserFindFromUserID(uid string) (*User, error) {
	return dm.userFind("uid", uid)
}

func (dm *DatabaseManager) UserCount() (int64, error) {
	var count int64

	err := dm.db.Model(&User{}).Count(&count).Error

	return count, err
}
