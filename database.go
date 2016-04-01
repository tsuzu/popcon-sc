package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// Shared in all codes
var mainDB *DatabaseManager

// DatabaseManager is a connector to this database
type DatabaseManager struct {
	db             *sql.DB
	showedNewCount int
}

// NewDatabaseManager is a function to initialize database connections
// static function
func NewDatabaseManager() (*DatabaseManager, error) {
	dm := &DatabaseManager{}
	var err error

	// pcpjudge Database
	dm.db, err = sql.Open("mysql", "popcon:password@/popcon") // Should change password

	if err != nil {
		return nil, err
	}

	dm.db.SetMaxIdleConns(150)

	err = dm.db.Ping()

	if err != nil {
		return nil, err
	}

	// user_and_group.go
	// Create Users Table
	_, err = dm.db.Exec("create table if not exists users (internalID int(11) auto_increment primary key, userID varchar(20) unique, userName varchar(256) unique, passHash varbinary(64), email varchar(50), groupID int(11))")

	if err != nil {
		return nil, err
	}

	// session.go
	// Create Sessions Table
	// TODO: Fix a bug about Year 2038 Bug in unixTimeLimit
	_, err = dm.db.Exec("create table if not exists sessions (sessionID varchar(50) primary key, internalID int(11), unixTimeLimit int(11), index iid(internalID), index idx(unixTimeLimit))")

	if err != nil {
		return nil, err
	}

	// user_and_group.go
	_, err = dm.db.Exec("create table if not exists groups (groupID int(11) auto_increment primary key, groupName varchar(50))")

	if err != nil {
		return nil, err
	}

	// news.go
	_, err = dm.db.Exec("create table if not exists news (text varchar(256), unixTime int, index uti(unixTime))")

	if err != nil {
		return nil, err
	}

	return dm, nil
}
