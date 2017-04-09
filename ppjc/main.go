package main

import (
	"os"

	"net/http"

	"strconv"

	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	muxlib "github.com/gorilla/mux"
)

func main() {
	InitLogger(os.Stdout, os.Getenv("PP_DEBUG_MODE") == "1")

	GeneralSetting.RankingRunningTerm = 10
	GeneralSetting.SavingTerm = 5

	//token := os.Getenv("POPCON_SC_RANKING_TOKEN")
	addr := os.Getenv("POPCON_SC_RANKING_ADDR")
	token := os.Getenv("PP_TOKEN")
	//	db := os.Getenv("POPCON_SC_RANKING_DB")

	mux := muxlib.NewRouter()

	mux.HandleFunc("/ranking/{cid}", func(rw http.ResponseWriter, req *http.Request) {
		vars := muxlib.Vars(req)

		cid, err := strconv.ParseInt(vars["cid"], 10, 64)

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)
		}
	})

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get("Authentication")

		if auth != token {
			respondJSON(rw, "Authentication Error", nil)

			return
		}

		mux.ServeHTTP(rw, req)
	})

	if err := http.ListenAndServe(":80", handler); err != nil {
		HttpLog().WithError(err).Fatal("ListenAndServe() error")

		return
	}

	return
}
