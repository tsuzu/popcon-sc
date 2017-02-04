package main

import "github.com/cs3238-tsuzu/popcon-sc/types"
import "database/sql"

const HttpCookieSession = "popcon_session"

var Location = popconSCTypes.Location

var NullStringInvalid = sql.NullString{Valid: false}
