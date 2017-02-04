package main

import "net/http"

// ParseRequestForSession returns a SessionTemplateData object
// error: ErrUnknownSession or ...
func ParseRequestForSession(req *http.Request) (*SessionTemplateData, error) {
	session := ParseSession(req)

	if session == nil {
		return nil, ErrUnknownSession
	}

	t, err := GetSessionTemplateData(*session)

	if err == ErrUnknownUser {
		return nil, ErrUnknownSession
	}

	return t, err
}

//ParseRequestForUserData returns an User object
// error: ErrUnknownSession or ...
func ParseRequestForUserData(req *http.Request) (*User, error) {
	sessionID := ParseSession(req)

	if sessionID == nil {
		return nil, ErrUnknownSession
	}

	u, err := GetSessionUserData(*sessionID)

	if err == ErrUnknownUser {
		return nil, ErrUnknownSession
	}

	return u, err
}

// ParseSession gets session from Cookie
func ParseSession(req *http.Request) *string {
	cookies := req.Cookies()
	var session *string

	for idx := range cookies {
		if cookies[idx].Name == HttpCookieSession {
			session = &cookies[idx].Value
		}
	}

	return session
}
