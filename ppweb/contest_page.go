package main

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"
	"path"

	"context"

	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/gorilla/mux"
)

func mustParseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)

	if err != nil {
		panic(err)
	}

	return u
}

const ContentsPerPage = 50

type ContestsTopContextKeyType int

const (
	ContestsTopContextKey ContestsTopContextKeyType = iota
)

type ContestsTopHandler struct {
	Temp        *template.Template
	NewContest  *template.Template
	EachHandler *ContestEachHandler
	Router      *mux.Router
}

func CreateContestsTopHandler() (*ContestsTopHandler, error) {
	funcs := template.FuncMap{
		"add": func(x, y int) int { return x + y },
		"TimeRangeToStringInt64": TimeRangeToStringInt64,
		"timeRangeToString":      func(st, fn time.Time) string { return TimeRangeToStringInt64(st.Unix(), fn.Unix()) },
		"contestTypeToString": func(t sctypes.ContestType) string {
			return sctypes.ContestTypeToString[t]
		},
	}

	temp, err := template.New("").Funcs(funcs).ParseFiles("./html/contests/index_tmpl.html")

	if err != nil {
		return nil, err
	}

	temp = temp.Lookup("index_tmpl.html")

	if temp == nil {
		return nil, errors.New("Failed to load /contests/index_temp.html")
	}

	newContest, err := template.ParseFiles("./html/contests/contest_new_tmpl.html")

	if err != nil {
		return nil, err
	}

	ceh, err := CreateContestEachHandler()

	if err != nil {
		return nil, err
	}

	router := mux.NewRouter()
	ch := &ContestsTopHandler{temp, newContest, ceh, router}

	router.HandleFunc("/{status:(?:|coming|closed)}", func(rw http.ResponseWriter, req *http.Request) {
		std := req.Context().Value(ContestsTopContextKey).(database.SessionTemplateData)
		status := mux.Vars(req)["status"]

		var cond []interface{}
		timeNow := time.Now()
		var reqType int
		if status == "" {
			reqType = 0
			cond = ArgumentsToArray("start_time<=? and finish_time>?", timeNow, timeNow)
		} else if status == "coming" {
			reqType = 1
			cond = ArgumentsToArray("start_time>?", timeNow)
		} else {
			reqType = 2
			cond = ArgumentsToArray("finish_time<=?", timeNow)
		}

		wrapForm := createWrapFormInt64(req)

		page := int(wrapForm("p"))

		if page == -1 {
			page = 1
		}

		count, err := mainDB.ContestCount(cond)

		if err != nil {
			HttpLog().Println(std.Iid, err)
			return
		}

		canCreateContest, err := mainRM.CanCreateContest()

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)
			DBLog().WithError(err).Error("CanCreateContest() error")

			return
		}

		type TemplateVal struct {
			Contests         []database.Contest
			UserName         string
			Type             int
			Current          int
			MaxPage          int
			Pagination       []PaginationHelper
			CanCreateContest bool
		}
		var templateVal TemplateVal
		templateVal.UserName = std.UserName
		templateVal.Type = reqType
		templateVal.Current = 1
		templateVal.CanCreateContest = (canCreateContest || std.Gid == 0)

		templateVal.MaxPage = int(count) / ContentsPerPage

		if int(count)%ContentsPerPage != 0 {
			templateVal.MaxPage++
		} else if templateVal.MaxPage == 0 {
			templateVal.MaxPage = 1
		}

		if count > 0 {
			if (page-1)*ContentsPerPage > int(count) {
				page = 1
			}

			templateVal.Current = page

			contests, err := mainDB.ContestList((page-1)*ContentsPerPage, ContentsPerPage, cond)

			if err == nil {
				templateVal.Contests = *contests
			} else {
				DBLog().WithError(err).WithField("iid", std.Iid).Error("ContestList error")
			}
		}

		templateVal.Pagination = NewPaginationHelper(templateVal.Current, templateVal.MaxPage, 3)

		rw.WriteHeader(http.StatusOK)
		ch.Temp.Execute(rw, templateVal)
	})
	router.HandleFunc("/new", func(rw http.ResponseWriter, req *http.Request) {
		std := req.Context().Value(ContestsTopContextKey).(database.SessionTemplateData)
		ch.newContestHandler(rw, req, std)
	})
	router.PathPrefix("/{cid:[0-9]+}/").HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		cid_ := mux.Vars(req)["cid"]
		cid, _ := strconv.ParseInt(cid_, 10, 64)
		std := req.Context().Value(ContestsTopContextKey).(database.SessionTemplateData)

		req, err := ch.EachHandler.PrepareVariables(req, cid, std)

		if err != nil {
			HttpLog().WithError(err).Error("PrepareVariables() error")
			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

			return
		}

		http.StripPrefix("/"+cid_, ch.EachHandler).ServeHTTP(rw, req)
	})

	return ch, nil
}

func TimeRangeToStringInt64(start, finish int64) string {
	startTime := time.Unix(start, 0)
	finishTime := time.Unix(finish, 0)

	return startTime.In(Location).Format("2006/01/02 15:04:05") + "-" + finishTime.In(Location).Format("2006/01/02 15:04:05")
}

func (ch ContestsTopHandler) newContestHandler(rw http.ResponseWriter, req *http.Request, std database.SessionTemplateData) {
	type TemplateVal struct {
		UserName       string
		Msg            string
		StartDate      string
		StartTime      string
		FinishDate     string
		FinishTime     string
		Description    string
		ContestName    string
		ContestTypes   map[sctypes.ContestType]string
		ContestTypeStr string
		Penalty        int64
	}

	canCreateContest, err := mainRM.CanCreateContest()

	if err != nil {
		sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)
		DBLog().WithError(err).Error("CanCreateContest() error")

		return
	}

	if !canCreateContest && std.Gid != 0 {
		sctypes.ResponseTemplateWrite(http.StatusForbidden, rw)

		return
	}

	if req.Method == "POST" {
		wrapFormStr := createWrapFormStr(req)
		wrapFormInt64 := createWrapFormInt64(req)

		startDate, startTime := wrapFormStr("start_date"), wrapFormStr("start_time")
		finishDate, finishTime := wrapFormStr("finish_date"), wrapFormStr("finish_time")

		description := wrapFormStr("description")
		contestName := wrapFormStr("contest_name")
		contestTypeStr := wrapFormStr("contest_type")
		penalty := wrapFormInt64("penalty")
		startStr := startDate + " " + startTime
		finishStr := finishDate + " " + finishTime

		var contestType sctypes.ContestType
		if !func() bool {
			for k, v := range sctypes.ContestTypeToString {
				if v == contestTypeStr {
					contestType = k
					return true
				}
			}
			return false
		}() {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}

		if penalty < 0 || penalty > 10000 {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}

		if len(contestName) == 0 || !UTF8StringLengthAndBOMCheck(contestName, 40) || strings.TrimSpace(contestName) == "" {
			msg := "コンテスト名が不正です。"
			templateVal := TemplateVal{
				std.UserName, msg, startDate, startTime, finishDate, finishTime, description, contestName, sctypes.ContestTypeToString, contestTypeStr, penalty,
			}

			ch.NewContest.Execute(rw, templateVal)

			return
		}

		start, err := time.ParseInLocation("2006/01/02 15:04", startStr, Location)

		if err != nil {
			msg := "開始日時の値が不正です。"
			templateVal := TemplateVal{
				std.UserName, msg, startDate, startTime, finishDate, finishTime, description, contestName, sctypes.ContestTypeToString, contestTypeStr, penalty,
			}

			ch.NewContest.Execute(rw, templateVal)

			return
		}

		finish, err := time.ParseInLocation("2006/01/02 15:04", finishStr, Location)

		if err != nil {
			msg := "終了日時の値が不正です。"
			templateVal := TemplateVal{
				std.UserName, msg, startDate, startTime, finishDate, finishTime, description, contestName, sctypes.ContestTypeToString, contestTypeStr, penalty,
			}

			ch.NewContest.Execute(rw, templateVal)

			return
		}

		if start.Unix() >= finish.Unix() || start.Unix() < time.Now().Unix() {
			msg := "開始日時または終了日時の値が不正です。"
			templateVal := TemplateVal{
				std.UserName, msg, startDate, startTime, finishDate, finishTime, description, contestName, sctypes.ContestTypeToString, contestTypeStr, penalty,
			}

			ch.NewContest.Execute(rw, templateVal)

			return
		}

		ppjcNew := func(cid int64) error {
			return ppjcClient.ContestsNew(cid)
		}

		cid, err := mainDB.ContestAdd(contestName, start, finish, std.Iid, contestType, penalty, ppjcNew)

		if err != nil {
			if strings.Index(err.Error(), "Duplicate") != -1 {
				msg := "すでに存在するコンテスト名です。"
				templateVal := TemplateVal{
					std.UserName, msg, startDate, startTime, finishDate, finishTime, description, contestName, sctypes.ContestTypeToString, contestTypeStr, penalty,
				}

				ch.NewContest.Execute(rw, templateVal)

				return
			} else {
				DBLog().WithError(err).Error("ContestAdd error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}
		}

		err = (&database.Contest{
			Cid: cid,
		}).DescriptionUpdate(description)

		if err != nil {
			HttpLog().WithError(err).WithField("cid", cid).Error("DescriptionUpdate() error")
		}

		RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/")
	} else if req.Method == "GET" {
		templateVal := TemplateVal{
			UserName:     std.UserName,
			ContestTypes: sctypes.ContestTypeToString,
		}

		ch.NewContest.Execute(rw, templateVal)
	} else {
		sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

		return
	}
}

func (ch ContestsTopHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	std, err := ParseRequestForSession(req)

	if err != nil {
		if err == database.ErrUnknownSession {
			RespondRedirection(rw, "/login?comeback=" + url.QueryEscape(path.Join("/contests", req.URL.Path)))

			return
		}

		DBLog().WithError(err).Error("ParseRequestForSession() error")

		sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

		return
	}

	req = req.WithContext(context.WithValue(req.Context(), ContestsTopContextKey, *std))

	err = req.ParseForm()

	if err != nil {
		sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

		return
	}

	ch.Router.ServeHTTP(rw, req)
}
