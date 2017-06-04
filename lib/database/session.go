package database

import "encoding/hex"
import "github.com/satori/go.uuid"
import "time"

// SessionTemplateData contains data in SQL DB
//"create table if not exists sessions (sessionKey varchar(50) primary key, internalID int(11), unixTimeLimit int(11), index iid(internalID), index idx(unixTimeLimit))"
type Session struct {
	SessionKey string `gorm:"primary_key;size:50"`
	Iid        int64  `gorm:"not null;index"`
	TimeLimit  int64  `gorm:"not null;index"`
}

type SessionTemplateData struct {
	Iid      int64
	UserID   string
	UserName string
	Gid      int64
}

func (dm *DatabaseManager) CreateSessionTable() error {
	return dm.BeginDM(func(dm *DatabaseManager) error {
		err := dm.db.AutoMigrate(&Session{}).Error

		if err != nil && !IsAlreadyExistsError(err) {
			return err
		}

		return nil
	})
}

// GetSessionTemplateData returns a SessionTemplateData object
func GetSessionTemplateData(sessionKey string) (*SessionTemplateData, error) {
	user, err := GetSessionUserData(sessionKey)

	if err != nil {
		return nil, err
	}

	return &SessionTemplateData{user.Iid, user.Uid, user.UserName, user.Gid}, nil
}

// GetSessionUserData returns an User object
func GetSessionUserData(sessionID string) (*User, error) {
	session, err := mainDB.SessionFind(sessionID)

	if err != nil {
		return nil, err
	}

	return mainDB.UserFindFromIID(session.Iid)
}

// SessionAdd adds a new session
func (dm *DatabaseManager) SessionAdd(internalID int64) (string, error) {
	var err error

	cnt := 0
	for {
		u := uuid.NewV4()
		id := hex.EncodeToString(u[:])
		session := Session{id, internalID, time.Now().Unix() + int64(720*time.Hour)}

		err = dm.db.Create(&session).Error

		if err == nil {
			return id, nil
		}

		if cnt > 3 {
			break
		}
		cnt++
	}

	return "", err
}

// SessionFind is to find a session
// len(sessionID) = 32
func (dm *DatabaseManager) SessionFind(sessionKey string) (*Session, error) {
	var results []Session
	err := dm.db.Where("session_key=?", sessionKey).Find(&results).Error

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ErrUnknownSession
	}

	return &results[0], nil
}

// SessionRemove is to remove session
// len(sessionKey) = 32
func (dm *DatabaseManager) SessionRemove(sessionKey string) error {
	return dm.db.Delete(&Session{SessionKey: sessionKey}).Error
}

//TODO implement caches of sessions
