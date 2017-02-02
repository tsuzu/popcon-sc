package main

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/naoina/genmai"
	//	"os"
)

// Shared in all codes
var mainDB *DatabaseManager

// DatabaseManager is a connector to this database
type DatabaseManager struct {
	db *genmai.DB
}

// NewDatabaseManager is a function to initialize database connections
// static function
func NewDatabaseManager(wait bool) (*DatabaseManager, error) {
	dm := &DatabaseManager{}
	var err error
	cnt := 0
	const RetyingMax = 100

RETRY:
	if cnt != 0 {
		DBLog.Info("Waiting for MySQL Server Launching...", err.Error())
		time.Sleep(3 * time.Second)
	}
	cnt++

	// Database
	dm.db, err = genmai.New(&genmai.MySQLDialect{}, settingManager.Get().dbAddr)

	if err != nil {
		if cnt > RetyingMax {
			return nil, err
		}

		goto RETRY
	}

	dm.db.DB().SetConnMaxLifetime(3 * time.Minute)
	dm.db.DB().SetMaxIdleConns(150)
	dm.db.DB().SetMaxOpenConns(150)

	err = dm.db.DB().Ping()

	if err != nil {
		if cnt > RetyingMax {
			return nil, err
		}

		goto RETRY
	}

	// user_and_group.go
	// Create Users Table
	err = dm.CreateUserTable()

	if err != nil {
		return nil, err
	}

	// session.go
	// Create Sessions Table
	err = dm.CreateSessionTable()

	if err != nil {
		return nil, err
	}

	// user_and_group.go
	err = dm.CreateGroupTable()

	if err != nil {
		return nil, err
	}

	// news.go
	err = dm.CreateNewsTable()

	if err != nil {
		return nil, err
	}

	err = dm.CreateContestTable()

	if err != nil {
		return nil, err
	}

	err = dm.CreateContestProblemTable()

	if err != nil {
		return nil, err
	}

	err = dm.CreateSubmissionTable()

	if err != nil {
		return nil, err
	}

	err = dm.CreateContestParticipationTable()

	if err != nil {
		return nil, err
	}

	err = dm.CreateLanguageTable()

	if err != nil {
		return nil, err
	}

	return dm, nil
}
