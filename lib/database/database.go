package database

import (
	"errors"
	"time"

	// DBの共通化のため
	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	// genmaiのサポートが危うくなっているため規模の大きいgormに移行

	"fmt"

	"github.com/cs3238-tsuzu/popcon-sc/lib/filesystem"
	"github.com/cs3238-tsuzu/popcon-sc/lib/redis"
	"github.com/jinzhu/gorm"
)

// DatabaseManager is a connector to this database
type DatabaseManager struct {
	db                 *gorm.DB
	fs                 *fs.MongoFSManager
	redis              *redis.RedisManager
	logger             func() *logrus.Entry
	transactionStarted bool
}

func (dm *DatabaseManager) Close() {
	dm.db.Close()
}

// NewDatabaseManager is a function to initialize database connections
// static function
func NewDatabaseManager(addr string, debugMode bool, fs *fs.MongoFSManager, redis *redis.RedisManager, logger func() *logrus.Entry) (*DatabaseManager, error) {
	specifyError := func(cat string, err error) error {
		return errors.New("In " + cat + ", " + err.Error())
	}

	dm := &DatabaseManager{
		redis:  redis,
		fs:     fs,
		logger: logger,
	}

	var err error
	cnt := 0
	const RetryingMax = 1000

RETRY:
	if cnt != 0 {
		logger().Info("Waiting for MySQL Server Launching...", err.Error())
		time.Sleep(3 * time.Second)
	}
	cnt++

	// Database
	dm.db, err = gorm.Open("mysql", addr)

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

	return dm, err
}

func (dm *DatabaseManager) BeginDM(f func(dm *DatabaseManager) error) (ret error) {
	if dm.transactionStarted {
		return ErrAlreadyTransactionBegun
	}
	db := dm.db.Begin()

	if db.Error != nil {
		return db.Error
	}

	clone := dm.Clone(db)
	clone.transactionStarted = true

	defer func() {
		if err := recover(); err != nil {
			db.Rollback()
			e, ok := err.(error)

			if ok {
				ret = e
			} else {
				ret = errors.New(fmt.Sprint(err))
			}
		}
	}()

	err := f(clone)

	if err != nil {
		db.Rollback()
		return err
	}

	err = db.Commit().Error

	return err
}

func (dm *DatabaseManager) BeginDMIfNotStarted(f func(dm *DatabaseManager) error) error {
	if err := dm.BeginDM(f); err != nil {
		if err == ErrAlreadyTransactionBegun {
			return f(dm)
		}

		return err
	}

	return nil
}

func (dm *DatabaseManager) Begin(f func(db *gorm.DB) error) error {
	return dm.BeginDM(func(dm *DatabaseManager) error {
		return f(dm.db)
	})
}

func (dm *DatabaseManager) BeginIfNotStarted(f func(dm *gorm.DB) error) error {
	if err := dm.Begin(f); err != nil {
		if err == ErrAlreadyTransactionBegun {
			return f(dm.db)
		}

		return err
	}

	return nil
}

func (dm *DatabaseManager) Logger() *logrus.Entry {
	return dm.logger()
}

func (dm *DatabaseManager) Clone(db *gorm.DB) *DatabaseManager {
	return &DatabaseManager{
		db:     db,
		fs:     dm.fs,
		redis:  dm.redis,
		logger: dm.logger,
	}
}

func (dm *DatabaseManager) DB() *gorm.DB {
	return dm.db
}
