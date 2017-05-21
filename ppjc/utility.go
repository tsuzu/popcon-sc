package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func GetCurrentUnixTime() int64 {
	return time.Now().Unix()
}

// RespondedJSON respondJson()向け
type RespondedJSON struct {
	Error   *string     `json:"error"`
	Message interface{} `json:"message"`
}

func stringToPointer(str string) *string {
	if str == "" {
		return nil
	}

	return &str
}

// respondJSON REST API向けにJSON返却をラップ
// エラー時でも200 OKを返すことに注意。
func respondJSON(rw http.ResponseWriter, errorMessage string, v interface{}) error {
	b, err := json.Marshal(RespondedJSON{Error: stringToPointer(errorMessage), Message: v})

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)

		return err
	}

	rw.Header()["Content-Type"] = []string{"application/json"}
	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write(b)

	return err
}

// parseForm POSTされたFormを取り出し
func parseForm(req *http.Request, formName string) (string, error) {
	err := req.ParseForm()

	if err != nil {
		return "", err
	}

	return req.Form.Get(formName), nil
}

// parseFormInt POSTされたFormを取り出し
func parseFormInt(req *http.Request, formName string) (int, error) {
	p, err := parseForm(req, formName)

	if err != nil {
		return 0, err
	}

	q, err := strconv.ParseInt(p, 10, 32)

	return int(q), err
}

// parseFormInt64 POSTされたFormを取り出し
func parseFormInt64(req *http.Request, formName string) (int64, error) {
	p, err := parseForm(req, formName)

	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(p, 10, 64)
}
