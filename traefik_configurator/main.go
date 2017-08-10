package main

import "fmt"
import "github.com/gorilla/mux"
import "net/http"

const AuthHeader = "Popcon-Authentication"

func main() {
	auth := os.Getenv("PPTC_AUTH")

	router := mux.NewRouter()

	http.HandlerFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get(AuthHeader) != auth {

		}

	})
	http.ListenAndServe(":22222")
}
