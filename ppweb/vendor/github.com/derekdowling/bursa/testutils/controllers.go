package testutils

// This file adds some nice helpers for performing controller tests

import (
	"log"
	"net/http"
	"net/url"
)

// Formats a post form submission url
func urlForm(path string, form url.Values) string {
	url := url.URL{
		Host:     "localhost:8080",
		Path:     path,
		RawQuery: form.Encode(),
	}

	return url.String()
}

// Builds a nice post form submissions request you can use in testing
func FormPostRequest(path string, form url.Values) (*http.Request, error) {
	url := urlForm(path, form)
	req, err := http.NewRequest("POST", url, nil)

	if err != nil {
		log.Print(err.Error())
	}

	return req, err
}
