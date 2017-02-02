package app

// This handles rendering our unauthenticated user facing static web pages.

import (
	"net/http"

	"github.com/derekdowling/bursa/picasso"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	// Temporary command to get the ball rolling
	picasso.Render(w, "app/layout", "app/index", nil)
}
