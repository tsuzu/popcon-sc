package main

import "errors"
import "crypto/sha512"
import "strconv"
import "database/sql"

// User is a struct to save UserData
// "create table if not exists users (internalID int(11) auto_increment primary key, uid varchar(20) unique, userName varchar(256) unique, passHash varbinary(64), email varchar(50), groupID int(11))"
type User struct {
	Iid      int64  `db:"pk"`
	Uid      string `default:"" size:"128"`
	UserName string `default:"" size:"256"`
	PassHash []byte
	Email    sql.NullString `default:"" size:"128"`
	Gid      int64          `default:""`
	Enabled  bool           `default:"true"`
}

func (dm *DatabaseManager) CreateUserTable() error {
	err := dm.db.CreateTableIfNotExists(&User{})

	if err != nil {
		return err
	}

	dm.db.CreateUniqueIndex(&User{}, "uid")
	dm.db.CreateUniqueIndex(&User{}, "user_name")
	dm.db.CreateUniqueIndex(&User{}, "email")

	return err
}

// UserAdd adds a new user
func (dm *DatabaseManager) UserAdd(uid string, userName string, pass string, email sql.NullString, groupID int64, enabled bool) (int64, error) {
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

	res, err := dm.db.DB().Exec("insert into user (uid, user_name, pass_hash, email, gid, enabled) values (?, ?, ?, ?, ?, ?)", uid, userName, passHashArr[:], email, groupID, enabled)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

// UserUpdate is a function to add a new user
// len(uid) <= 20, len(userName) <= 256 len(pass) <= 50, len(email) <= 50
func (dm *DatabaseManager) UserUpdate(internalID int, uid string, userName string, pass string, email string, groupID int64, enabled bool) error {
	if len(uid) > 128 {
		return errors.New("error: len(uid) > 128")
	}

	if len(userName) > 256 {
		return errors.New("error: len(userName) > 256")
	}

	if len(pass) > 100 {
		return errors.New("error: len(pass) > 50")
	}
	passHashArr := sha512.Sum512([]byte(pass))

	if len(email) > 128 {
		return errors.New("error: len(email) > 128")
	}

	_, err := dm.db.Update(&User{
		Uid:      uid,
		UserName: userName,
		PassHash: passHashArr[:],
		Email:    NullStringCreate(email),
		Gid:      groupID,
		Enabled:  enabled,
	})

	return err
}

func (dm *DatabaseManager) UserUpdateEnabled(iid int64, enabled bool) error {
	_, err := dm.db.DB().Exec("update user set enabled=? where iid=?", enabled, iid)

	return err
}

// UserFind returns a User object
func (dm *DatabaseManager) userFind(key string, value string) (*User, error) {
	var resulsts []User
	err := dm.db.Select(&resulsts, dm.db.Where(key, "=", value))

	if err != nil {
		return nil, err
	}

	if len(resulsts) == 0 {
		return nil, errors.New("Unknown user")
	}

	return &resulsts[0], nil
}

// UserFindFromIID returns a User object
func (dm *DatabaseManager) UserFindFromIID(internalID int64) (*User, error) {
	return dm.userFind("iid", strconv.FormatInt(internalID, 10))
}

// UserFindFromuid returns a User object
func (dm *DatabaseManager) UserFindFromUserID(uid string) (*User, error) {
	if len(uid) > 20 {
		return nil, errors.New("error: len(uid) > 20")
	}

	return dm.userFind("uid", uid)

}

func (dm *DatabaseManager) UserCount() (int64, error) {
	var count int64

	err := dm.db.Select(&count, dm.db.Count("iid"), dm.db.From(&User{}))

	return count, err
}
