package database

import (
	"testing"

	"github.com/Sirupsen/logrus"
)

func newDBM() (*DatabaseManager, error) {
	return NewDatabaseManager("root:test@tcp(mysql:3306)/popcon?charset=utf8mb4&parseTime=True", true, nil, nil, func() *logrus.Entry { return logrus.NewEntry(logrus.StandardLogger()) })
}

func TestRankingAutoMigrate(t *testing.T) {

	dm, err := newDBM()

	if err != nil {
		t.Fatal(err)
	}

	if err := dm.RankingAutoMigrate(1, []int64{1, 2, 3, 5}); err != nil {
		t.Fatal(err)
	}

	if err := dm.RankingUserAdd(1, 10); err != nil {
		t.Fatal(err)
	}
}
func TestRankingUserAdd(t *testing.T) {

	dm, err := newDBM()

	if err != nil {
		t.Fatal(err)
	}

	if err := dm.RankingUserAdd(1, 10); err != nil {
		t.Fatal(err)
	}

}
