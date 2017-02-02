package store

import (
	"log"

	"github.com/derekdowling/bursa/models"
	_ "github.com/lib/pq"
)

// Having trouble with this. The hope was that init would create our database
// connection for us and that we can share it subsequently. It doesn't seem to
// quite work:
//    Line 47: - runtime error: invalid memory address or nil pointer dereference
//    goroutine 10 [running]:
//    runtime.panic(0x69ed20, 0xa1cd88)
//      /usr/local/go/src/pkg/runtime/panic.c:248 +0x106
//    github.com/jinzhu/gorm.(*Scope).fieldFromStruct(0xc21006f500, 0x6ef100, 0x2, 0x0, 0x0, ...)
// var db gorm.DB

// Because of the private model this is just an inefficient way of boostrapping
// and ensuring our model table exists without forcing us to run a migration task.
// TODO do this properly as some kind of migration command? Or maybe this *is*
// the way to do it in go land.
func init() {
	db, db_err := models.Connect()
	if db_err != nil {
		log.Fatalf("Failed to connect to key storage.", db_err)
	}

	// TODO error check here
	models.Initialize()

	if db_err := db.CreateTable(UserKey{}).Error; db_err != nil {
		// Gorm already logs for us.
	}
}

// Private model struct for storing User -> Private key mappings in postgres.
// Eventually this might be moved to some more secure system where a "heavy"
// package like postgres might not be available so we hide this and just code against
// a Store / Retrieve API.
type UserKey struct {
	Id     int64
	UserId int64
	// Well - multiple users might have access to a given key. Lets not worry about
	// that for now. Or never. If you've got two people that have a given key just
	// make them use a single account? Or does that putz with some our ideas?
	ExtendedKey string `sql:"type:varchar(255);"`
}

// On the topic of user_id instead of username - You know, I like the idea of
// usernames TBH. Surrogate keys are ugly.  If I'm looking at raw database
// results it's way more useful to see jacob_str than fucking 123100sdfasdfa.
// But - it seems it's always tempting to start using them though the reasons
// never make sense to me aftewards. Usernames should be unique and unchanging.
func Store(user_id int64, encoded_base_58_key string) error {
	// See comment for the global db variable at the start of this file.
	db, db_err := models.Connect()
	if db_err != nil {
		log.Fatalf("Failed to connect to key storage.", db_err)
	}

	user_key := UserKey{
		UserId:      user_id,
		ExtendedKey: encoded_base_58_key,
	}

	if db_err := db.Save(&user_key).Error; db_err != nil {
		return db_err
	} else {
		return nil
	}
}

// Retrieve returns an encoded_base_58_key for a given user_id. Currently this is
// going to be broken because a user might have multiple private keys and this
// only returns the first. A common use case is likely going to be filtering by
// a public key corresponding to one (or none) of the user's private keys.
func Retrieve(user_id int64) (string, error) {
	// See comment for the global db variable at the start of this file.
	db, db_err := models.Connect()
	if db_err != nil {
		log.Fatalf("Failed to connect to key storage.", db_err)
	}

	var user_key UserKey
	// TODO error checking goes here!
	db.Where("user_id = ?", user_id).First(&user_key)
	return user_key.ExtendedKey, nil
}
