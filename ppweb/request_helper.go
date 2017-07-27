package main

import (
	"net/http"

	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
)

// ParseRequestForSession returns a SessionTemplateData object
// error: ErrUnknownSession or ...
func ParseRequestForSession(req *http.Request) (*database.SessionTemplateData, error) {
	session := ParseSession(req)

	if session == nil {
		return nil, database.ErrUnknownSession
	}

	t, err := mainDB.GetSessionTemplateData(*session)

	if err == database.ErrUnknownUser {
		return nil, database.ErrUnknownSession
	}

	return t, err
}

//ParseRequestForUserData returns an User object
// error: ErrUnknownSession or ...
func ParseRequestForUserData(req *http.Request) (*database.User, error) {
	sessionID := ParseSession(req)

	if sessionID == nil {
		return nil, database.ErrUnknownSession
	}

	u, err := mainDB.GetSessionUserData(*sessionID)

	if err == database.ErrUnknownUser {
		return nil, database.ErrUnknownSession
	}

	return u, err
}

// ParseSession gets session from Cookie
func ParseSession(req *http.Request) *string {
	cookies := req.Cookies()
	var session *string

	for idx := range cookies {
		if cookies[idx].Name == HTTPCookieSession {
			session = &cookies[idx].Value
		}
	}

	return session
}
