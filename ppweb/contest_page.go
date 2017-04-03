package main

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/cs3238-tsuzu/popcon-sc/types"
)

const ContentsPerPage = 50

type ContestsTopHandler struct {
	Temp        *template.Template
	NewContest  *template.Template
	EachHandler *ContestEachHandler
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

	return &ContestsTopHandler{temp, newContest, ceh}, nil
}

func TimeRangeToStringInt64(start, finish int64) string {
	startTime := time.Unix(start, 0)
	finishTime := time.Unix(finish, 0)

	return startTime.In(Location).Format("2006/01/02 15:04:05") + "-" + finishTime.In(Location).Format("2006/01/02 15:04:05")
}

func (ch ContestsTopHandler) newContestHandler(rw http.ResponseWriter, req *http.Request, std *SessionTemplateData) {
	type TemplateVal struct {
		UserName    string
		Msg         *string
		StartDate   string
		StartTime   string
		FinishDate  string
		FinishTime  string
		Description string
		ContestName string
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

		startDate, startTime := wrapFormStr("start_date"), wrapFormStr("start_time")
		finishDate, finishTime := wrapFormStr("finish_date"), wrapFormStr("finish_time")
		description := wrapFormStr("description")
		contestName := wrapFormStr("contest_name")

		startStr := startDate + " " + startTime
		finishStr := finishDate + " " + finishTime

		if len(contestName) == 0 || !UTF8StringLengthAndBOMCheck(contestName, 40) || strings.TrimSpace(contestName) == "" {
			msg := "コンテスト名が不正です。"
			templateVal := TemplateVal{
				std.UserName, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
			}

			ch.NewContest.Execute(rw, templateVal)

			return
		}

		start, err := time.ParseInLocation("2006/01/02 15:04", startStr, Location)

		if err != nil {
			msg := "開始日時の値が不正です。"
			templateVal := TemplateVal{
				std.UserName, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
			}

			ch.NewContest.Execute(rw, templateVal)

			return
		}

		finish, err := time.ParseInLocation("2006/01/02 15:04", finishStr, Location)

		if err != nil {
			msg := "終了日時の値が不正です。"
			templateVal := TemplateVal{
				std.UserName, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
			}

			ch.NewContest.Execute(rw, templateVal)

			return
		}

		if start.Unix() >= finish.Unix() || start.Unix() < time.Now().Unix() {
			msg := "開始日時または終了日時の値が不正です。"
			templateVal := TemplateVal{
				std.UserName, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
			}

			ch.NewContest.Execute(rw, templateVal)

			return
		}

		cid, err := mainDB.ContestAdd(contestName, start, finish, std.Iid, 0)

		if err != nil {
			if strings.Index(err.Error(), "Duplicate") != -1 {
				msg := "すでに存在するコンテスト名です。"
				templateVal := TemplateVal{
					std.UserName, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
				}

				ch.NewContest.Execute(rw, templateVal)

				return
			} else {
				DBLog().WithError(err).Error("ContestAdd error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}
		}

		err = (&Contest{
			Cid: cid,
		}).DescriptionUpdate(description)

		if err != nil {
			HttpLog().Println(std.Iid, err)
		}

		RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/")
	} else if req.Method == "GET" {
		templateVal := TemplateVal{
			UserName: std.UserName,
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
		if err == ErrUnknownSession {
			RespondRedirection(rw, path.Join("/login?comeback=/contests", url.QueryEscape(req.URL.Path)))

			return
		}

		DBLog().WithError(err).Error("ParseRequestForSession failed")

		sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

		return
	}

	err = req.ParseForm()

	if err != nil {
		sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

		return
	}

	var cond []interface{}
	timeNow := time.Now()
	var reqType int

	switch req.URL.Path {
	case "/":
		reqType = 0
		cond = ArgumentsToArray("start_time<=? and finish_time>?", timeNow, timeNow)
	case "/coming/":
		reqType = 1
		cond = ArgumentsToArray("start_time>?", timeNow)
	case "/closed/":
		reqType = 2
		cond = ArgumentsToArray("finish_time<=?", timeNow)
	case "/new":
		ch.newContestHandler(rw, req, std)

		return
	default:
		if len(req.URL.Path) == 0 {
			RespondRedirection(rw, "/contests/")

			return
		}

		idx := strings.Index(req.URL.Path[1:], "/")

		if idx == -1 {
			RespondRedirection(rw, "/contests"+req.URL.Path+"/")

			return
		}

		cidStr := req.URL.Path[1:][:idx]

		cid, err := strconv.ParseInt(cidStr, 10, 64)

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

			return
		}

		handler, err := ch.EachHandler.GetHandler(cid, *std)

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

			return
		}

		http.StripPrefix("/"+cidStr, handler).ServeHTTP(rw, req)

		return
	}

	wrapForm := createWrapForm(req)

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
		Contests         []Contest
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
			HttpLog().WithError(err).WithField("iid", std.Iid).Error("ContestList error")
		}
	}

	templateVal.Pagination = NewPaginationHelper(templateVal.Current, templateVal.MaxPage, 3)

	rw.WriteHeader(http.StatusOK)
	ch.Temp.Execute(rw, templateVal)

}
