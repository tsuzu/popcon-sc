// This handles rendering all of our unauthenticated user facing static web pages
package styleguide

import (
	"github.com/derekdowling/bursa/picasso"
	"net/http"
)

// Completes a user signup. Assumes that the values being provided from the
// front-end have already been validated
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	picasso.Render(w, "style-guide/layout", "style-guide/index", nil)
}

