package main

import (
	"context"
	htmlTemplate "html/template"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"fmt"

	"io"

	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

type ContestEachHandler struct {
	TopPage                      *template.Template
	ProblemList                  *template.Template
	ProblemView                  *template.Template
	SubmissionList               *template.Template
	SubmissionView               *template.Template
	SubmitPage                   *template.Template
	ManagementTopPage            *template.Template
	ManagementRejudgePage        *template.Template
	ManagementSettingPage        *template.Template
	ManagementProblemSettingPage *template.Template
	ManagementProblemList        *template.Template
	ManagementTastcaseList       *template.Template
	ManagementTestcaseSetting    *template.Template
	RankingPage                  *template.Template
	Router                       *mux.Router
}

type ContestEachContextKeyType int

const (
	ContestEachContextKey ContestEachContextKeyType = iota
)

type ContestEachPreparedData struct {
	Cid            int64
	Std            database.SessionTemplateData
	Contest        *database.Contest
	Joined         bool
	IsAdmin        bool
	IsSpecialAdmin bool
	IsStarted      bool
	IsFinished     bool
	Accessible     bool
}

func (ceh *ContestEachHandler) PrepareVariables(req *http.Request, cid int64, std database.SessionTemplateData) (*http.Request, error) {
	var err error
	var pdata ContestEachPreparedData

	pdata.Cid = cid
	pdata.Contest, err = mainDB.ContestFind(cid)
	pdata.Std = std

	if err != nil {
		return nil, err
	}

	pdata.Joined, pdata.IsAdmin, err = mainDB.ContestParticipationCheck(std.Iid, cid)

	if err != nil {
		return nil, err
	}

	pdata.IsSpecialAdmin = (pdata.Contest.Admin == std.Iid)

	if pdata.IsSpecialAdmin {
		pdata.IsAdmin = true
	}

	now := time.Now().Unix()
	pdata.IsStarted = (pdata.Contest.StartTime.Unix() <= now)
	pdata.IsFinished = (pdata.Contest.FinishTime.Unix() <= now)

	pdata.Accessible = (pdata.Joined && pdata.IsStarted) || pdata.IsFinished || pdata.IsAdmin

	return req.WithContext(context.WithValue(req.Context(), ContestEachContextKey, pdata)), nil
}

func (ceh *ContestEachHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ceh.Router.ServeHTTP(rw, req)
}

func CreateContestEachHandler() (*ContestEachHandler, error) {
	funcMap := template.FuncMap{
		"timeRangeToStringInt64": TimeRangeToStringInt64,
	}

	top, err := template.New("").Funcs(funcMap).ParseFiles("./html/contests/each/index_tmpl.html")

	if err != nil {
		return nil, err
	}

	probList, err := template.ParseFiles("./html/contests/each/problems_tmpl.html")

	if err != nil {
		return nil, err
	}

	probView, err := template.ParseFiles("./html/contests/each/problem_view_tmpl.html")

	if err != nil {
		return nil, err
	}

	funcMap = template.FuncMap{
		"timeToString": TimeToString,
		"add":          func(x, y int) int { return x + y },
		"timeRangeConv": func(x, y int64) string {
			if y == 0 {
				return "00:00"
			}

			str := fmt.Sprintf("%02d", (y-x)/60%60) + ":" + fmt.Sprintf("%02d", (y-x)%60)

			if (y-x)/3600 != 0 {
				str = fmt.Sprintf("%02d", (y-x)/3600) + ":" + str
			}

			return str
		},
	}

	subList, err := template.New("").Funcs(funcMap).ParseFiles("./html/contests/each/submissions_tmpl.html")

	if err != nil {
		return nil, err
	}

	subView, err := template.New("").Funcs(funcMap).ParseFiles("./html/contests/each/submission_view_tmpl.html")

	if err != nil {
		return nil, err
	}

	submit, err := template.ParseFiles("./html/contests/each/submit_tmpl.html")

	if err != nil {
		return nil, err
	}

	man, err := template.ParseFiles("./html/contests/each/management_tmpl.html")

	if err != nil {
		return nil, err
	}

	manre, err := template.ParseFiles("./html/contests/each/management/rejudge_tmpl.html")

	if err != nil {
		return nil, err
	}

	rank, err := template.New("").Funcs(funcMap).ParseFiles("./html/contests/each/ranking_tmpl.html")

	if err != nil {
		return nil, err
	}

	funcMap = template.FuncMap{
		"timeRangeToStringInt64": TimeRangeToStringInt64,
	}

	manse, err := template.New("").Funcs(funcMap).ParseFiles("./html/contests/each/management/setting_tmpl.html")

	if err != nil {
		return nil, err
	}

	manpr, err := template.ParseFiles("./html/contests/each/management/problem_set_tmpl.html")

	if err != nil {
		return nil, err
	}

	manprv, err := template.ParseFiles("./html/contests/each/management/problems_tmpl.html")

	if err != nil {
		return nil, err
	}

	mantc, err := template.ParseFiles("./html/contests/each/management/testcases_tmpl.html")

	if err != nil {
		return nil, err
	}

	mantcv, err := template.ParseFiles("./html/contests/each/management/testcase_each_tmpl.html")

	if err != nil {
		return nil, err
	}

	router := mux.NewRouter()
	ceh := &ContestEachHandler{
		top.Lookup("index_tmpl.html"),
		probList,
		probView,
		subList.Lookup("submissions_tmpl.html"),
		subView.Lookup("submission_view_tmpl.html"),
		submit,
		man,
		manre,
		manse.Lookup("setting_tmpl.html"),
		manpr,
		manprv,
		mantc,
		mantcv,
		rank.Lookup("ranking_tmpl.html"),
		router,
	}

	router.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		//TODO: Update not found page
		sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)
	})
	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)

		type TemplateVal struct {
			UserName               string
			Cid                    int64
			ContestName            string
			Description            htmlTemplate.HTML
			JoinButtonActive       bool
			StartTime              int64
			FinishTime             int64
			Enabled                bool
			ManagementButtonActive bool
		}

		desc, err := (&database.Contest{Cid: pdata.Cid}).DescriptionLoad()

		if err != nil {
			desc = ""
		}

		unsafe := blackfriday.MarkdownCommon([]byte(desc))

		policy := bluemonday.UGCPolicy()
		policy.AllowAttrs("width").OnElements("img")
		policy.AllowAttrs("height").OnElements("img")

		html := policy.SanitizeBytes(unsafe)

		templateVal := TemplateVal{
			UserName:               pdata.Std.UserName,
			Cid:                    pdata.Cid,
			ContestName:            pdata.Contest.Name,
			Description:            htmlTemplate.HTML(html),
			JoinButtonActive:       !(pdata.IsFinished || pdata.Joined || pdata.IsAdmin),
			StartTime:              pdata.Contest.StartTime.Unix(),
			FinishTime:             pdata.Contest.FinishTime.Unix(),
			Enabled:                pdata.Accessible,
			ManagementButtonActive: pdata.IsAdmin,
		}

		rw.WriteHeader(http.StatusOK)
		ceh.TopPage.Execute(rw, templateVal)
	})

	router.HandleFunc("/problems/", func(rw http.ResponseWriter, req *http.Request) {
		pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
		if !pdata.Accessible {
			RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/")

			return
		}

		probList, err := mainDB.ContestProblemList(pdata.Cid)

		if err != nil {
			probList = []database.ContestProblem{}
		}

		type TemplateVal struct {
			ContestName string
			Problems    []database.ContestProblem
			UserName    string
			Cid         int64
		}

		templateVal := TemplateVal{
			pdata.Contest.Name,
			probList,
			pdata.Std.UserName,
			pdata.Cid,
		}

		rw.WriteHeader(http.StatusOK)
		ceh.ProblemList.Execute(rw, templateVal)

		return
	})

	router.HandleFunc("/problems/{pidx:[0-9]+}", func(rw http.ResponseWriter, req *http.Request) {
		pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
		pidx, _ := strconv.ParseInt(mux.Vars(req)["pidx"], 10, 64)

		prob, err := mainDB.ContestProblemFind2(pdata.Cid, pidx)

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

			return
		}

		stat, err := prob.LoadStatement()

		if err != nil {
			DBLog().WithError(err).Error("LoadStatement error")
			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

			return
		}

		unsafe := blackfriday.MarkdownCommon([]byte(stat))
		html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

		type TemplateVal struct {
			database.ContestProblem
			ContestName string
			Cid         int64
			Text        string
			UserName    string
		}
		templateVal := TemplateVal{*prob, pdata.Contest.Name, pdata.Contest.Cid, string(html), pdata.Std.UserName}
		rw.WriteHeader(http.StatusOK)
		ceh.ProblemView.Execute(rw, templateVal)
	})

	router.HandleFunc("/ranking", func(rw http.ResponseWriter, req *http.Request) {
		pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
		if !pdata.Accessible {
			RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/")

			return
		}

		wrapForm := createWrapFormInt64(req)

		page := int(wrapForm("p"))

		if page == -1 {
			page = 1
		}

		type RankingRow2 struct {
			database.RankingRow
			Rank int
		}

		type TemplateVal struct {
			ContestName string
			Cid         int64
			UserName    string
			Problems    []database.ContestProblem
			Ranking     []RankingRow2
			Current     int
			MaxPage     int
			BeginTime   int64
			Pagination  []PaginationHelper
		}

		count := 0
		/*		count, err := mainDB.ContestRankingCount(pdata.Cid)

				if err != nil {
					DBLog().WithError(err).WithField("iid", pdata.Std.Iid).Error("ContestRankingCount error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}*/

		var templateVal TemplateVal
		templateVal.Cid = pdata.Cid
		templateVal.ContestName = pdata.Contest.Name
		templateVal.UserName = pdata.Std.UserName
		templateVal.BeginTime = pdata.Contest.StartTime.Unix()
		probs, err := mainDB.ContestProblemList(pdata.Cid)

		if err != nil {
			DBLog().WithError(err).WithField("iid", pdata.Std.Iid).Error("ContestProblemList error")
			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

			return
		}
		templateVal.Problems = probs
		templateVal.Current = 1
		templateVal.MaxPage = int(count) / ContentsPerPage

		if int(count)%ContentsPerPage != 0 {
			templateVal.MaxPage++
		} else if templateVal.MaxPage == 0 {
			templateVal.MaxPage = 1
		}

		if count > 0 {
			/*if (page-1)*ContentsPerPage > int(count) {
				page = 1
			}

			templateVal.Current = page

			ranks, err := mainDB.ContestRankingList(pdata.Cid, int64((page-1)*ContentsPerPage), ContentsPerPage)

			if err != nil {
				DBLog().WithError(err).Error("ContestRankingList error")

				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			ranks2 := make([]RankingRow2, len(ranks))

			for i := range ranks {
				ranks2[i] = RankingRow2{ranks[i], (page-1)*ContentsPerPage + i + 1}
			}

			templateVal.Ranking = ranks2*/
		}

		templateVal.Pagination = NewPaginationHelper(templateVal.Current, templateVal.MaxPage, 3)

		ceh.RankingPage.Execute(rw, templateVal)
	})

	router.HandleFunc("/submissions/", func(rw http.ResponseWriter, req *http.Request) {
		pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
		if !pdata.Accessible {
			RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/")

			return
		}

		wrapForm := createWrapFormInt64(req)

		wrapFormStr := createWrapFormStr(req)

		stat := wrapForm("status")
		lang := wrapForm("lang")
		prob := wrapForm("prob")
		page := int(wrapForm("p"))
		userID := wrapFormStr("user")

		const IllegalParam = -128
		if page == -1 {
			page = 1
		}

		var iid int64
		if userID == "" {
			iid = -1
		} else {
			if len(userID) > 40 || !UTF8StringLengthAndBOMCheck(userID, 40) {
				sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

				return
			}

			user, err := mainDB.UserFindFromUserID(userID)

			if err != nil {
				iid = IllegalParam
			} else {
				iid = user.Iid
			}
		}

		if !(pdata.IsFinished || pdata.IsAdmin) && iid != pdata.Std.Iid {
			RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/submissions/?user="+pdata.Std.UserID)

			return
		}

		count, err := mainDB.SubmissionViewCount(pdata.Cid, iid, lang, prob, stat)

		if err != nil {
			DBLog().WithError(err).WithField("iid", pdata.Std.Iid).Error("SubmissionViewCount error")
			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

			return
		}

		type TemplateVal struct {
			AllEnabled  bool
			ContestName string
			UserName    string
			Cid         int64
			Uid         string
			Submissions []database.SubmissionView
			Problems    []database.ContestProblem
			Languages   []database.Language
			Current     int
			MaxPage     int
			Pagination  []PaginationHelper
			Lang        int64
			Prob        int64
			Status      int64
			User        string
		}
		var templateVal TemplateVal
		templateVal.AllEnabled = pdata.IsFinished || pdata.IsAdmin
		templateVal.ContestName = pdata.Contest.Name
		templateVal.Cid = pdata.Cid
		templateVal.UserName = pdata.Std.UserName
		templateVal.User = userID
		templateVal.Status = stat
		templateVal.Lang = lang
		templateVal.Prob = prob
		templateVal.Uid = pdata.Std.UserID

		langs, err := mainDB.LanguageActiveList()

		if err != nil {
			HttpLog().WithField("iid", iid).WithError(err).Error("LanguageList() error")
		} else {
			templateVal.Languages = langs
		}

		probs, err := mainDB.ContestProblemListLight(pdata.Cid)

		if err != nil {
			HttpLog().WithField("iid", iid).WithError(err).Error("ContestProblemListLight() error")
		} else {
			templateVal.Problems = probs
		}

		templateVal.Current = 1

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

			submissions, err := mainDB.SubmissionViewList(pdata.Cid, iid, lang, prob, stat, int64((page-1)*ContentsPerPage), ContentsPerPage)

			if err == nil {
				templateVal.Submissions = submissions
			} else {
				HttpLog().WithField("iid", iid).WithError(err).Error("SubmissionViewList() error")
			}
		}

		templateVal.Pagination = NewPaginationHelper(templateVal.Current, templateVal.MaxPage, 3)

		rw.WriteHeader(200)

		ceh.SubmissionList.Execute(rw, templateVal)

	})

	router.HandleFunc("/submissions/{sid:[0-9]+}", func(rw http.ResponseWriter, req *http.Request) {
		pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
		sid, _ := strconv.ParseInt(mux.Vars(req)["sid"], 10, 64)

		submission, err := mainDB.SubmissionViewFind(sid, pdata.Cid)

		if err == database.ErrUnknownSubmission {
			sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

			return
		} else if err != nil {
			DBLog().WithError(err).WithField("sid", sid).Error("SubmissionViewFind error")
			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

			return
		}

		if !pdata.IsAdmin && submission.Iid != pdata.Std.Iid && !pdata.IsFinished {
			sctypes.ResponseTemplateWrite(http.StatusForbidden, rw)

			return
		}

		code, err := mainDB.SubmissionGetCode(pdata.Cid, sid)

		if err != nil {
			var tmp string

			code = tmp
		}

		type SubmissionTestCaseView struct {
			database.SubmissionTestCase
			StatusString string
		}

		casesArr, err := mainDB.SubmissionGetCase(pdata.Cid, sid)
		var caseViews []SubmissionTestCaseView

		if err == nil {
			caseViews = make([]SubmissionTestCaseView, 0, len(casesArr))
			for _, v := range casesArr {
				caseViews = append(caseViews, SubmissionTestCaseView{v, sctypes.SubmissionStatusTypeToString[v.Status]})
			}
		} else {
			HttpLog().WithError(err).Error("SubmissionGetCase() error")
		}

		msg, err := mainDB.SubmissionGetMsg(pdata.Cid, sid)

		if err != nil {
			msg = ""
		}

		type TemplateVal struct {
			ContestName string
			Submission  database.SubmissionView
			Cases       []SubmissionTestCaseView
			Code        string
			Msg         string
			UserName    string
			Cid         int64
		}

		templateVal := TemplateVal{
			ContestName: pdata.Contest.Name,
			Submission:  *submission,
			Cases:       caseViews,
			Code:        code,
			Msg:         msg,
			UserName:    pdata.Std.UserName,
			Cid:         pdata.Cid,
		}

		rw.WriteHeader(http.StatusOK)
		ceh.SubmissionView.Execute(rw, templateVal)
	})

	router.HandleFunc("/join", func(rw http.ResponseWriter, req *http.Request) {
		pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)

		if req.Method == "GET" && !pdata.IsAdmin && !pdata.Accessible {
			mainDB.ContestParticipationAdd(pdata.Std.Iid, pdata.Cid)

			RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/")
		} else {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}
	})

	router.HandleFunc("/submit", func(rw http.ResponseWriter, req *http.Request) {
		pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
		if !pdata.Accessible {
			RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/")

			return
		}

		if req.Method == "GET" {
			type TemplateVal struct {
				ContestName string
				UserName    string
				Cid         int64
				Problems    []database.ContestProblem
				Languages   []database.Language
				Prob        int64
			}

			list, err := mainDB.ContestProblemListLight(pdata.Cid)

			if err != nil {
				list = []database.ContestProblem{}

				HttpLog().WithError(err).Error("ContestProblemListLight() error")
			}

			lang, err := mainDB.LanguageActiveList()

			if err != nil {
				lang = []database.Language{}

				HttpLog().WithError(err).Error("LanguageList() error")
			}

			probArr, has := req.Form["prob"]
			var prob int64 = -1

			if has && len(probArr) != 0 {
				p, err := strconv.ParseInt(probArr[0], 10, 64)

				if err != nil {
					prob = -1
				}
				prob = p
			}

			templateVal := TemplateVal{
				pdata.Contest.Name,
				pdata.Std.UserName,
				pdata.Cid,
				list,
				lang,
				prob,
			}

			rw.WriteHeader(http.StatusOK)
			ceh.SubmitPage.Execute(rw, templateVal)
		} else if req.Method == "POST" {
			wrapForm := createWrapFormInt64(req)

			wrapFormStr := createWrapFormStr(req)

			lid := wrapForm("lang")
			pid := wrapForm("prob")
			code := wrapFormStr("code")

			if lid < 0 || pid < 0 || code == "" {
				sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

				return
			}

			prob, err := mainDB.ContestProblemFind2(pdata.Cid, pid)

			if err != nil {
				if err == database.ErrUnknownProblem {
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				} else {
					DBLog().WithError(err).Error("ContestProblemFind2 error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}
			}

			_, err = mainDB.LanguageFind(lid)

			if err != nil {
				if err == database.ErrUnknownLanguage {
					rw.WriteHeader(http.StatusBadRequest)
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				} else {
					DBLog().WithError(err).Error("LanguageFind error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}
			}

			sid, err := mainDB.SubmissionAdd(pdata.Cid, prob.Pid, pdata.Std.Iid, lid, code)

			if err != nil {
				DBLog().WithError(err).Error("SubmissionAdd error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}
			if err := ppjcClient.JudgeSubmit(pdata.Cid, sid); err != nil {
				HttpLog().WithError(err).WithField("cid", pdata.Cid).WithField("sid", sid).Error("JudgeSubmit() error")

				mainDB.SubmissionUpdateResult(pdata.Cid, sid, 0, sctypes.SubmissionStatusInternalError, 0, strings.NewReader("internal server error(judge queue is down)"))
			}

			RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/submissions/"+strconv.FormatInt(sid, 10))
		} else {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)
		}
	})

	// /management/*
	func() {
		sub := mux.NewRouter()
		stripped := http.StripPrefix("/management", sub)
		router.PathPrefix("/management/").HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
			if !pdata.IsAdmin {
				sctypes.ResponseTemplateWrite(http.StatusForbidden, rw)

				return
			}

			stripped.ServeHTTP(rw, req)
		})

		sub.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
			pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)

			type TemplateVal struct {
				Cid         int64
				ContestName string
				UserName    string
			}
			ceh.ManagementTopPage.Execute(rw, TemplateVal{pdata.Cid, pdata.Std.UserName, pdata.Contest.Name})
		})

		sub.HandleFunc("/remove", func(rw http.ResponseWriter, req *http.Request) {
			// TODO: Remove ranking from ppjc
			pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)

			list, err := mainDB.ContestProblemList(pdata.Cid)

			if err != nil {
				DBLog().WithError(err).Error("ContestProblemList() error")
			}

			for i := range list {
				err = mainDB.SubmissionRemoveAll(pdata.Cid, list[i].Pid)

				if err != nil {
					DBLog().WithError(err).Error("SubmissionRemoveAll() error")
				}
			}

			err = mainDB.ContestParticipationRemove(pdata.Cid)

			if err != nil {
				DBLog().WithError(err).Error("ContestParticipationRemove() error")
			}

			err = mainDB.ContestDelete(pdata.Cid)

			if err != nil {
				DBLog().WithError(err).Error("ContestDelete() error")
			}

			RespondRedirection(rw, "/contests/")

			return
		})

		sub.HandleFunc("/rejudge", func(rw http.ResponseWriter, req *http.Request) {
			pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
			respondTemp := func(msg string) {
				type TemplateVal struct {
					Cid         int64
					UserName    string
					Msg         *string
					ContestName string
				}

				if msg == "" {
					ceh.ManagementRejudgePage.Execute(rw, TemplateVal{pdata.Cid, pdata.Std.UserName, nil, pdata.Contest.Name})
				} else {
					ceh.ManagementRejudgePage.Execute(rw, TemplateVal{pdata.Cid, pdata.Std.UserName, &msg, pdata.Contest.Name})
				}
			}

			if req.Method == "GET" {
				respondTemp("")

				return
			} else if req.Method == "POST" {
				wrapForm := createWrapFormInt64(req)

				target, id := wrapForm("target"), wrapForm("id")

				if (target != 1 && target != 2) || id < 0 {
					respondTemp("不正なIDです。")

					return
				}

				if target == 1 {
					sm, err := mainDB.SubmissionFind(pdata.Cid, id)

					if err != nil {
						if err == database.ErrUnknownSubmission {
							respondTemp("該当する提出がありません。")
						} else {
							DBLog().WithError(err).Error("SubmissionFind error")
							sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)
						}
						return
					}

					ppjcClient.JudgeSubmit(pdata.Cid, sm.Sid)

					RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/management/")

					return
				} else {
					cp, err := mainDB.ContestProblemFind2(pdata.Cid, id)

					if err != nil {
						if err == database.ErrUnknownProblem {
							respondTemp("該当する問題がありません。")
						} else {
							DBLog().WithError(err).Error("ContestProblemFind2 error")
							sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						}
						return
					}

					sml, err := mainDB.SubmissionListWithPid(pdata.Cid, cp.Pid)

					if err != nil {
						DBLog().WithError(err).Error("SubmissionList error")
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}

					for i := range sml {
						if err := ppjcClient.JudgeSubmit(pdata.Cid, sml[i].Sid); err != nil {
							HttpLog().WithError(err).WithField("cid", pdata.Cid).WithField("sid", sml[i].Sid).Error("JudgeSubmit() error")
						}
					}

					RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/management/")

					return
				}
			} else {
				sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

				return
			}
		})

		sub.HandleFunc("/setting", func(rw http.ResponseWriter, req *http.Request) {
			pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
			type TemplateVal struct {
				Cid         int64
				UserName    string
				Msg         *string
				StartDate   string
				StartTime   string
				FinishDate  string
				FinishTime  string
				Description string
				ContestName string
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
						pdata.Cid, pdata.Std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManagementSettingPage.Execute(rw, templateVal)

					return
				}

				start, err := time.ParseInLocation("2006/01/02 15:04", startStr, Location)

				if err != nil {
					msg := "開始日時の値が不正です。"
					templateVal := TemplateVal{
						pdata.Cid, pdata.Std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManagementSettingPage.Execute(rw, templateVal)

					return
				}

				if pdata.Contest.StartTime.Unix() <= time.Now().Add(2*time.Minute).Unix() && pdata.Contest.StartTime.Unix() != start.Unix() {
					msg := "開始日時は2分前を切ると変更できません。"

					startDate = pdata.Contest.StartTime.In(Location).Format("2006/01/02")
					startTime = pdata.Contest.StartTime.In(Location).Format("15:04")

					templateVal := TemplateVal{
						pdata.Cid, pdata.Std.UserID, &msg, startDate, pdata.Contest.StartTime.In(Location).Format("2006/01/02 15:04"), finishDate, finishTime, description, contestName,
					}

					ceh.ManagementSettingPage.Execute(rw, templateVal)

					return
				}

				finish, err := time.ParseInLocation("2006/01/02 15:04", finishStr, Location)

				if err != nil {
					msg := "終了日時の値が不正です。"
					templateVal := TemplateVal{
						pdata.Cid, pdata.Std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManagementSettingPage.Execute(rw, templateVal)

					return
				}

				if pdata.Contest.FinishTime.Unix() <= time.Now().Add(2*time.Minute).Unix() && pdata.Contest.FinishTime.Unix() != finish.Unix() {
					msg := "終了日時は2分前を切ると変更できません。"

					finishDate = pdata.Contest.FinishTime.In(Location).Format("2006/01/02")
					finishTime = pdata.Contest.FinishTime.In(Location).Format("15:04")

					templateVal := TemplateVal{
						pdata.Cid, pdata.Std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManagementSettingPage.Execute(rw, templateVal)

					return
				}

				if start.Unix() >= finish.Unix() || (pdata.Contest.StartTime.Unix() != start.Unix() && start.Unix() < time.Now().Unix()) || (pdata.Contest.FinishTime.Unix() != finish.Unix() && finish.Unix() < time.Now().Unix()) {
					msg := "開始日時及び終了日時の値が不正です。"
					templateVal := TemplateVal{
						pdata.Cid, pdata.Std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManagementSettingPage.Execute(rw, templateVal)

					return
				}

				err = mainDB.ContestUpdate(pdata.Cid, contestName, start, finish, pdata.Contest.Admin, 0)

				if err != nil {
					if strings.Index(err.Error(), "Duplicate") != -1 {
						msg := "すでに存在するコンテスト名です。"
						templateVal := TemplateVal{
							pdata.Cid, pdata.Std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
						}

						ceh.ManagementSettingPage.Execute(rw, templateVal)

						return
					} else {
						DBLog().WithError(err).Error("ContestUpdate error")
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}
				}

				err = (&database.Contest{Cid: pdata.Cid}).DescriptionUpdate(description)

				if err != nil {
					HttpLog().WithError(err).Error("DescriptionUpdate() error")
				}

				RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/management/")
			} else if req.Method == "GET" {
				desc, _ := (&database.Contest{Cid: pdata.Cid}).DescriptionLoad()

				templateVal := TemplateVal{
					Cid:         pdata.Cid,
					UserName:    pdata.Std.UserID,
					StartDate:   pdata.Contest.StartTime.In(Location).Format("2006/01/02"),
					StartTime:   pdata.Contest.StartTime.In(Location).Format("15:04"),
					FinishDate:  pdata.Contest.FinishTime.In(Location).Format("2006/01/02"),
					FinishTime:  pdata.Contest.FinishTime.In(Location).Format("15:04"),
					ContestName: pdata.Contest.Name,
					Description: desc,
				}
				ceh.ManagementSettingPage.Execute(rw, templateVal)
			} else {
				sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

				return
			}
		})

		sub.HandleFunc("/problems/", func(rw http.ResponseWriter, req *http.Request) {
			pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)

			type TemplateVal struct {
				Cid         int64
				ContestName string
				UserName    string
				Problems    []database.ContestProblem
			}

			list, err := mainDB.ContestProblemList(pdata.Cid)

			if err != nil {
				DBLog().WithError(err).Error("ContestProblemList error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			ceh.ManagementProblemList.Execute(rw, TemplateVal{pdata.Cid, pdata.Contest.Name, pdata.Std.UserName, list})
		})
		sub.HandleFunc("/problems/new", func(rw http.ResponseWriter, req *http.Request) {
			pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)

			cnt, err := mainDB.ContestProblemCount(pdata.Cid)

			if err != nil {
				DBLog().WithError(err).Error("ContestProblemCount error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			type TemplateVal struct {
				Cid         int64
				ContestName string
				UserName    string
				Msg         string
				Mode        bool
				Pidx        int64
				Name        string
				Time        int64
				Mem         int64
				Type        int64
				Prob        string
				Lang        int64
				Languages   []database.Language
				Code        string
			}

			wrapForm := createWrapFormInt64(req)
			wrapFormStr := createWrapFormStr(req)
			languages, err := mainDB.LanguageList()

			if err != nil {
				DBLog().WithError(err).Error("LanguageList error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			var cp *database.ContestProblem

			if req.Method == "GET" {
				temp := TemplateVal{
					Cid:         pdata.Cid,
					ContestName: pdata.Contest.Name,
					Time:        1,
					Mem:         32,
					UserName:    pdata.Std.UserName,
					Mode:        true,
					Languages:   languages,
				}

				// TODO: Add the setting of the maximum number of problems
				if cnt >= 50 {
					temp.Msg = "コンテストの問題数の上限に達しているため新しく問題を追加することができません。"
				}

				ceh.ManagementProblemSettingPage.Execute(rw, temp)

				return
			} else if req.Method == "POST" {
				pidx, name, time, mem := wrapForm("pidx"), wrapFormStr("problem_name"), wrapForm("time"), wrapForm("mem")
				jtype, prob, lid, code := wrapForm("type"), wrapFormStr("prob"), wrapForm("lang"), wrapFormStr("code")

				if pidx == -1 || time < 1 || time > 10 || mem < 32 || mem > 1024 || jtype < 0 || jtype > 1 || (jtype == int64(sctypes.JudgeRunningCode) && lid == -1) {
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				}

				if len(name) == 0 || !UTF8StringLengthAndBOMCheck(name, 40) || strings.TrimSpace(name) == "" {
					ceh.ManagementProblemSettingPage.Execute(
						rw,
						TemplateVal{
							pdata.Cid,
							pdata.Contest.Name,
							pdata.Std.UserName,
							"問題名が不正です。",
							true,
							pidx,
							name,
							time,
							mem,
							jtype,
							prob,
							lid,
							languages,
							code,
						},
					)

					return
				}

				if cnt >= 50 {
					ceh.ManagementProblemSettingPage.Execute(
						rw,
						TemplateVal{
							pdata.Cid,
							pdata.Contest.Name,
							pdata.Std.UserName,
							"コンテストの問題数の上限に達しているため新しく問題を追加することができません。",
							true,
							pidx,
							name,
							time,
							mem,
							jtype,
							prob,
							lid,
							languages,
							code,
						},
					)
				}

				if sctypes.JudgeType(jtype) == sctypes.JudgeRunningCode {
					if _, err := mainDB.LanguageFind(lid); err != nil {
						if err == database.ErrUnknownLanguage {
							sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

							return
						} else {
							DBLog().WithError(err).Error("LanguageFind error")
							sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

							return
						}
					}
				}

				cp, err = pdata.Contest.ProblemAdd(pidx, name, time, mem, sctypes.JudgeType(jtype))

				if err == nil {
					if err := ppjcClient.ContestsProblemsAdd(pdata.Cid, cp.Pid); err != nil {
						HttpLog().WithError(err).WithField("pdata.Cid", pdata.Cid).WithField("pdata.Cid", cp.Pid).Error("ppjc.ContestProblemAdd() error")
						if err := mainDB.ContestProblemDelete(pdata.Cid, cp.Pid); err != nil {
							DBLog().WithError(err).WithField("pdata.Cid", pdata.Cid).WithField("pid", cp.Pid).Error("ContestProblemDelete() error")
						}
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}
				}

				if err != nil {
					if database.IsDuplicateError(err) {
						ceh.ManagementProblemSettingPage.Execute(
							rw,
							TemplateVal{
								pdata.Cid,
								pdata.Contest.Name,
								pdata.Std.UserName,
								"使用されている問題番号です。",
								true,
								pidx,
								name,
								time,
								mem,
								jtype,
								prob,
								lid,
								languages,
								code,
							},
						)

						return
					} else {
						DBLog().WithError(err).Error("ProblemAdd/ContestProblemUpdate error")
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}
				}

				err = cp.UpdateStatement(prob)

				if err != nil {
					DBLog().WithError(err).Error("UpdateStatement error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}

				err = cp.UpdateChecker(lid, code)

				if err != nil {
					DBLog().WithError(err).Error("UpdateChecker error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}

				RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/management/problems/")
			} else {
				sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

				return
			}

		})
		sub.HandleFunc("/problems/{upidx:[0-9]+}", func(rw http.ResponseWriter, req *http.Request) {
			pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
			upidx, _ := strconv.ParseInt(mux.Vars(req)["upidx"], 10, 64)

			cnt, err := mainDB.ContestProblemCount(pdata.Cid)

			if err != nil {
				DBLog().WithError(err).Error("ContestProblemCount error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			type TemplateVal struct {
				Cid         int64
				ContestName string
				UserName    string
				Msg         string
				Mode        bool
				Pidx        int64
				Name        string
				Time        int64
				Mem         int64
				Type        int64
				Prob        string
				Lang        int64
				Languages   []database.Language
				Code        string
			}

			wrapForm := createWrapFormInt64(req)
			wrapFormStr := createWrapFormStr(req)
			languages, err := mainDB.LanguageList()

			if err != nil {
				DBLog().WithError(err).Error("LanguageList error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			cp, err := mainDB.ContestProblemFind2(pdata.Cid, upidx)

			if err != nil {
				if err == database.ErrUnknownProblem {
					sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

					return
				} else {
					DBLog().WithError(err).WithField("cid", pdata.Cid).Error("ContestProblemFind2() error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}
			}

			if req.Method == "GET" {
				lid, checker, err := cp.LoadChecker()

				if err != nil {
					DBLog().WithError(err).Error("LoadChecker error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}

				stat, err := cp.LoadStatement()

				if err != nil {
					DBLog().WithError(err).Error("LoadStatement error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}
				temp := TemplateVal{
					Cid:         pdata.Cid,
					ContestName: pdata.Contest.Name,
					UserName:    pdata.Std.UserName,
					Mode:        false,
					Languages:   languages,
					Name:        cp.Name,
					Time:        cp.Time,
					Mem:         cp.Mem,
					Pidx:        upidx,
					Type:        int64(cp.Type),
					Lang:        lid,
					Code:        checker,
					Prob:        stat,
				}

				ceh.ManagementProblemSettingPage.Execute(rw, temp)

				return
			} else if req.Method == "POST" {
				pidx, name, time, mem := wrapForm("pidx"), wrapFormStr("problem_name"), wrapForm("time"), wrapForm("mem")
				jtype, prob, lid, code := wrapForm("type"), wrapFormStr("prob"), wrapForm("lang"), wrapFormStr("code")

				if pidx == -1 || time < 1 || time > 10 || mem < 32 || mem > 1024 || jtype < 0 || jtype > 1 || (jtype == int64(sctypes.JudgeRunningCode) && lid == -1) {
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				}

				if len(name) == 0 || !UTF8StringLengthAndBOMCheck(name, 40) || strings.TrimSpace(name) == "" {
					ceh.ManagementProblemSettingPage.Execute(
						rw,
						TemplateVal{
							pdata.Cid,
							pdata.Contest.Name,
							pdata.Std.UserName,
							"問題名が不正です。",
							true,
							pidx,
							name,
							time,
							mem,
							jtype,
							prob,
							lid,
							languages,
							code,
						},
					)

					return
				}

				if cnt >= 50 {
					ceh.ManagementProblemSettingPage.Execute(
						rw,
						TemplateVal{
							pdata.Cid,
							pdata.Contest.Name,
							pdata.Std.UserName,
							"コンテストの問題数の上限に達しているため新しく問題を追加することができません。",
							true,
							pidx,
							name,
							time,
							mem,
							jtype,
							prob,
							lid,
							languages,
							code,
						},
					)
				}

				if sctypes.JudgeType(jtype) == sctypes.JudgeRunningCode {
					if _, err := mainDB.LanguageFind(lid); err != nil {
						if err == database.ErrUnknownLanguage {
							sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

							return
						} else {
							DBLog().WithError(err).Error("LanguageFind error")
							sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

							return
						}
					}
				}

				cp.Pidx = pidx
				cp.Name = name
				cp.Time = time
				cp.Mem = mem
				cp.Type = sctypes.JudgeType(jtype)

				err = mainDB.ContestProblemUpdate(*cp)

				if err != nil {
					if database.IsDuplicateError(err) {
						ceh.ManagementProblemSettingPage.Execute(
							rw,
							TemplateVal{
								pdata.Cid,
								pdata.Contest.Name,
								pdata.Std.UserName,
								"使用されている問題番号です。",
								true,
								pidx,
								name,
								time,
								mem,
								jtype,
								prob,
								lid,
								languages,
								code,
							},
						)

						return
					} else {
						DBLog().WithError(err).Error("ProblemAdd/ContestProblemUpdate error")
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}
				}

				err = cp.UpdateStatement(prob)

				if err != nil {
					DBLog().WithError(err).Error("UpdateStatement error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}

				err = cp.UpdateChecker(lid, code)

				if err != nil {
					DBLog().WithError(err).Error("UpdateChecker error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}

				RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/management/problems/")
			} else {
				sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

				return
			}
		})
		sub.HandleFunc("/testcases/{pidx:[0-9]+}", func(rw http.ResponseWriter, req *http.Request) {
			pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
			pidx, _ := strconv.ParseInt(mux.Vars(req)["pidx"], 10, 64)

			cp, err := mainDB.ContestProblemFind2(pdata.Cid, pidx)

			if err == database.ErrUnknownProblem {
				sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

				return
			} else if err != nil {
				DBLog().WithError(err).Error("ContestProblemFind2() error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			type TemplateVal struct {
				Cid         int64
				Pidx        int64
				ContestName string
				UserName    string
				Testcases   []string
				Scoresets   []database.ContestProblemScoreSet
				Msg         string
			}

			if req.Method == "GET" {
				cases, sets, err := cp.LoadTestCaseNames()

				if err != nil {
					DBLog().WithError(err).Error("LoadTestCaseNames error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}

				ceh.ManagementTastcaseList.Execute(
					rw,
					TemplateVal{
						pdata.Cid,
						pidx,
						pdata.Contest.Name,
						pdata.Std.UserName,
						cases,
						sets,
						"",
					},
				)
			} else if req.Method == "POST" {
				caseNames := req.Form["case_name[]"]
				setScores := req.Form["set_score[]"]
				setCases := req.Form["set_case[]"]

				if len(caseNames) > 50 {
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				}

				if len(setScores) != len(setCases) || len(setScores) > 50 {
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				}

				cases := make([]string, len(caseNames))
				for i := range cases {
					cases[i] = caseNames[i]
				}
				illegal := false

				scores := make([]database.ContestProblemScoreSet, len(setScores))
				for i := range scores {
					caseIds := make([]int64, 0, 50)
					for _, str := range strings.Split(setCases[i], ",") {
						str = strings.TrimSpace(str)

						id, err := strconv.ParseInt(str, 10, 64)

						if err != nil {
							illegal = true
						}

						if id < 0 || int(id) >= len(cases) {
							illegal = true
						}

						caseIds = append(caseIds, id)
					}

					score, err := strconv.ParseInt(setScores[i], 10, 32)

					if err != nil {
						illegal = true
					}

					if score < 0 || score > 10000 {
						illegal = true
					}

					scores[i] = database.ContestProblemScoreSet{
						Score: score,
					}

					scores[i].Cases.Set(caseIds)
					scores[i].BeforeSave() // copy from cases to casesrawstring
				}

				if illegal {
					ceh.ManagementTastcaseList.Execute(
						rw,
						TemplateVal{
							pdata.Cid,
							pidx,
							pdata.Contest.Name,
							pdata.Std.UserName,
							cases,
							scores,
							"不正なパラメータがあります。",
						},
					)

					return
				}

				err := cp.UpdateTestCaseNames(cases, scores)

				if err != nil {
					DBLog().WithError(err).Error("UpdateTestCaseNames error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}

				RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/management/testcases/"+strconv.FormatInt(pidx, 10))
			} else {
				sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

				return
			}
		})
		sub.HandleFunc("/testcases/{pidx:[0-9]+}/{tcid:[0-9]+}", func(rw http.ResponseWriter, req *http.Request) {
			pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
			pidx, _ := strconv.ParseInt(mux.Vars(req)["pidx"], 10, 64)
			tcid, _ := strconv.ParseInt(mux.Vars(req)["tcid"], 10, 64)

			cp, err := mainDB.ContestProblemFind2(pdata.Cid, pidx)

			if err == database.ErrUnknownProblem {
				sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

				return
			} else if err != nil {
				DBLog().WithError(err).Error("ContestProblemFind2() error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			in, out, err := cp.LoadTestCaseInfo(tcid)

			if err == database.ErrUnknownTestcase {
				sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

				return
			} else if err != nil {
				DBLog().WithError(err).Error("LoadTestCaseInfo error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			type TemplateVal struct {
				Cid                     int64
				Id                      int
				UserName                string
				Pidx                    int64
				ProbName                string
				InCapacity, OutCapacity int64
			}

			templateVal := TemplateVal{
				pdata.Cid,
				int(tcid),
				pdata.Std.UserName,
				pidx,
				cp.Name,
				in, out,
			}

			ceh.ManagementTestcaseSetting.Execute(rw, templateVal)
		})
		sub.HandleFunc("/testcases/{pidx:[0-9]+}/{tcid:[0-9]+}/{mode:(?:input|output)}", func(rw http.ResponseWriter, req *http.Request) {
			pdata := req.Context().Value(ContestEachContextKey).(ContestEachPreparedData)
			vars := mux.Vars(req)
			pidx, _ := strconv.ParseInt(vars["pidx"], 10, 64)
			tcid, _ := strconv.ParseInt(vars["tcid"], 10, 64)
			mode := vars["mode"]

			cp, err := mainDB.ContestProblemFind2(pdata.Cid, pidx)

			if err == database.ErrUnknownProblem {
				sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

				return
			} else if err != nil {
				DBLog().WithError(err).Error("ContestProblemFind2() error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			if req.Method == "POST" {
				err := req.ParseMultipartForm(10 * 1024 * 1024)

				if err != nil {
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				}

				file, _, err := req.FormFile("file")

				if err != nil {
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				}
				l, err := file.Seek(0, 2)

				if err != nil {
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				} else if l > 100*1024*1024 {
					sctypes.ResponseTemplateWrite(http.StatusRequestEntityTooLarge, rw)

					return
				}

				file.Seek(0, 0)

				defer file.Close()

				if mode == "input" {
					err = cp.UpdateTestCase(true, tcid, NewTrimNewlineReader(file))
				} else {
					err = cp.UpdateTestCase(false, tcid, NewTrimNewlineReader(file))
				}

				if err == database.ErrUnknownTestcase {
					sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

					return
				} else if err != nil {
					DBLog().WithError(err).Error("UpdateTestCase error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}

				RespondRedirection(rw, "/contests/"+strconv.FormatInt(pdata.Cid, 10)+"/management/testcases/"+strconv.FormatInt(pidx, 10)+"/"+strconv.FormatInt(int64(tcid), 10))
			} else if req.Method == "GET" {
				var reader io.ReadCloser
				var err error

				if mode == "input" {
					reader, err = cp.LoadTestCase(true, int(tcid))
				} else {
					reader, err = cp.LoadTestCase(false, int(tcid))
				}

				if err == database.ErrUnknownTestcase {
					sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

					return
				} else if err != nil {
					DBLog().WithError(err).Error("UpdateTestCase error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}
				defer reader.Close()

				fileName := strconv.FormatInt(pdata.Cid, 10) + "-" + strconv.FormatInt(pidx, 10) + "_" + strconv.FormatInt(int64(tcid), 10)

				if mode == "input" {
					fileName += "_in.txt"
				} else {
					fileName += "_out.txt"
				}

				rw.Header()["X-Content-Type-Options"] = []string{"nosniff"}
				rw.Header()["Content-Type"] = []string{"text/plain; charset=UTF-8"}
				rw.Header()["Content-Disposition"] = []string{"attachment; filename=\"" + fileName + "\""}

				rw.WriteHeader(http.StatusOK)
				io.Copy(rw, reader)
			} else {
				sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

				return
			}
		})
	}()

	return ceh, nil
}
