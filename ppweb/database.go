package main

import (
	"errors"
	"time"

	_ "github.com/go-sql-driver/mysql"
	// genmaiのサポートが危うくなっているため規模の大きいgormに移行

	"fmt"

	"github.com/jinzhu/gorm"
)

// Shared in all codes
var mainDB *DatabaseManager

// DatabaseManager is a connector to this database
type DatabaseManager struct {
	db *gorm.DB
}

func (dm *DatabaseManager) Close() {
	dm.db.Close()
}

// NewDatabaseManager is a function to initialize database connections
// static function
func NewDatabaseManager(debugMode bool) (*DatabaseManager, error) {
	specifyError := func(cat string, err error) error {
		return errors.New("In " + cat + ", " + err.Error())
	}

	dm := &DatabaseManager{}
	var err error
	cnt := 0
	const RetryingMax = 1000

RETRY:
	if cnt != 0 {
		DBLog.Info("Waiting for MySQL Server Launching...", err.Error())
		time.Sleep(3 * time.Second)
	}
	cnt++

	// Database
	dm.db, err = gorm.Open("mysql", settingManager.Get().dbAddr)

	if err != nil {
		if cnt > RetryingMax {
			return nil, specifyError("connection", err)
		}

		goto RETRY
	}

	dm.db.DB().SetConnMaxLifetime(3 * time.Minute)
	dm.db.DB().SetMaxIdleConns(150)
	dm.db.DB().SetMaxOpenConns(150)
	if debugMode {
		dm.db.LogMode(true)
	}
	err = dm.db.DB().Ping()

	if err != nil {
		if cnt > RetryingMax {
			return nil, specifyError("connection", err)
		}

		dm.db.Close()
		goto RETRY
	}

	// user_and_group.go
	// Create Users Table
	err = dm.CreateUserTable()

	if err != nil {
		return nil, specifyError("user", err)
	}

	// session.go
	// Create Sessions Table
	err = dm.CreateSessionTable()

	if err != nil {
		return nil, specifyError("session", err)
	}

	// group.go
	err = dm.CreateGroupTable()

	if err != nil {
		return nil, specifyError("group", err)
	}

	// news.go
	err = dm.CreateNewsTable()

	if err != nil {
		return nil, specifyError("news", err)
	}

	err = dm.CreateContestTable()

	if err != nil {
		return nil, specifyError("contest", err)
	}

	err = dm.CreateContestProblemTable()

	if err != nil {
		return nil, specifyError("contest_problem", err)
	}

	err = dm.CreateSubmissionTable()

	if err != nil {
		return nil, specifyError("submission", err)
	}

	err = dm.CreateContestParticipationTable()

	if err != nil {
		return nil, specifyError("contest_participation", err)
	}

	err = dm.CreateLanguageTable()

	if err != nil {
		return nil, specifyError("language", err)
	}

	return dm, nil
}

func (dm *DatabaseManager) Begin(f func(*gorm.DB) error) error {
	db := dm.db.Begin()

	if db.Error != nil {
		return db.Error
	}

	err := f(db)

	if err != nil {
		db.Rollback()
		return err
	}

	if err := recover(); err != nil {
		db.Rollback()
		e, ok := err.(error)

		if ok {
			return e
		}

		return errors.New(fmt.Sprint(err))
	}

	err = db.Commit().Error

	return err
}
