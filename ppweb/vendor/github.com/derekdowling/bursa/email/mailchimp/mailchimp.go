package mailchimp

import (
	log "github.com/Sirupsen/logrus"
	"github.com/derekdowling/bursa/config"
	"github.com/mattbaird/gochimp"
)

// Adds a user, via their email, to one of our MailChimp mailing lists
func SubscribeToChimp(userEmail string) bool {
	chimp := getMailChimp()
	request := gochimp.ListsSubscribe{
		ListId:         getMailListId(),
		Email:          gochimp.Email{Email: userEmail},
		DoubleOptIn:    false,
		UpdateExisting: true,
		SendWelcome:    sendWelcomeEmail(),
	}

	_, err := chimp.ListsSubscribe(request)
	if err != nil {
		log.WithFields(log.Fields{
			"email": userEmail,
		}).Error(err.Error())
		return false
	}
	return true
}

// Checks whether or not we are in production to avoid spamming ourselves
// with email
func sendWelcomeEmail() bool {
	return config.App.GetBool("email.enabled")
}

// Sets up the MailChimp API
func getMailChimp() *gochimp.ChimpAPI {
	api_key := config.App.GetString("email.mailchimp_key")
	return gochimp.NewChimp(api_key, true)
}

// Determines which mailing list to add user to based on context
func getMailListId() string {
	return config.App.GetString("email.list_id")
}
