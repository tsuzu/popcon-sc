package models

import (
	"github.com/jinzhu/gorm"
	"time"

	"github.com/derekdowling/bursa/config"
	_ "github.com/lib/pq"
)

const (
	BitcoinCurrency = "bitcoin"
)

const (
	CompletedStatus  = "completed"
	PendingStatus    = "pending"
	FailedStatus     = "failed"
	RolledBackStatus = "rolledback"
)

const (
	CodeSuccess = 200
)

type Transfer struct {
	// TODO Refer to bitcoin protocol.
	Id string `sql:"type:varchar(255);unique"`

	// Amount in satoshis when using Bitcoin.
	Currency string `sql:"size:255"`
	Amount   int64

	SourceWallet      Wallet
	DestinationWallet Wallet

	CreatedBy   User
	CreatedAt   time.Time
	CompletedAt time.Time
	Status      string `sql:"size:32"`
	Code        int16  `sql:"size:8"`
}

func (self *Transfer) IsSuccess() bool {
	return self.Code == CodeSuccess
}

type Wallet struct {
	Id        string `sql:"type:varchar(255);unique"`
	Name      string `sql:"size:255"`
	Pin       string `sql:"size:255"`
	Balance   int64
	Transfers []Transfer

	CreatedAt time.Time
	CreatedBy User
	OwnedBy   User
}

// Well suited to some kind of management cli.
func Initialize() {
	db, _ := Connect()

	// See https://github.com/jinzhu/gorm/blob/master/migration_test.go#L23
	// Note that the scope of db_err is visible only to this if block.
	if db_err := db.CreateTable(User{}).Error; db_err != nil {
		// Gorm already logs for us.
	}

	if db_err := db.CreateTable(Transfer{}).Error; db_err != nil {
		// Gorm already logs for us.
	}

	if db_err := db.CreateTable(Wallet{}).Error; db_err != nil {
		// Gorm already logs for us.
	}
}

func Connect() (gorm.DB, error) {
	return gorm.Open(
		config.App.GetString("orm.adapter"),
		config.App.GetString("orm.settings"),
	)
}
