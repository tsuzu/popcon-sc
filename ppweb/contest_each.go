package main

import (
	htmlTemplate "html/template"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"fmt"

	"io"

	"github.com/cs3238-tsuzu/popcon-sc/types"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

type ContestEachHandler struct {
	// Template
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
}

func (ceh *ContestEachHandler) checkAdmin(cont *Contest, std SessionTemplateData) bool {
	if std.Gid == 0 {
		return true
	}

	if cont.Admin == std.Iid {
		return true
	}

	return false
}

func (ceh *ContestEachHandler) GetHandler(cid int64, std SessionTemplateData) (http.HandlerFunc, error) {
	cont, err := mainDB.ContestFind(cid)

	if err != nil {
		return nil, err
	}

	check, err := mainDB.ContestParticipationCheck(std.Iid, cid)

	if err != nil {
		return nil, err
	}

	isStarted := (cont.StartTime.Unix() <= time.Now().Unix())
	isFinished := (cont.FinishTime.Unix() <= time.Now().Unix())

	free := (check && isStarted) || isFinished

	isAdmin := ceh.checkAdmin(cont, std)

	if isAdmin {
		free = true
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

			return
		}

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

		desc, err := (&Contest{Cid: cid}).DescriptionLoad()

		if err != nil {
			desc = ""
		}

		unsafe := blackfriday.MarkdownCommon([]byte(desc))

		policy := bluemonday.UGCPolicy()
		policy.AllowAttrs("width").OnElements("img")
		policy.AllowAttrs("height").OnElements("img")

		html := policy.SanitizeBytes(unsafe)

		templateVal := TemplateVal{
			UserName:               std.UserName,
			Cid:                    cid,
			ContestName:            cont.Name,
			Description:            htmlTemplate.HTML(html),
			JoinButtonActive:       !(isFinished || check || isAdmin),
			StartTime:              cont.StartTime.Unix(),
			FinishTime:             cont.FinishTime.Unix(),
			Enabled:                free,
			ManagementButtonActive: isAdmin,
		}

		rw.WriteHeader(http.StatusOK)
		ceh.TopPage.Execute(rw, templateVal)
	})

	mux.HandleFunc("/problems/", func(rw http.ResponseWriter, req *http.Request) {
		if !free {
			RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/")

			return
		}

		http.StripPrefix("/problems/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "" {
				probList, err := mainDB.ContestProblemList(cid)

				if err != nil {
					probList = []ContestProblem{}
				}

				type TemplateVal struct {
					ContestName string
					Problems    []ContestProblem
					UserName    string
					Cid         int64
				}

				templateVal := TemplateVal{
					cont.Name,
					probList,
					std.UserName,
					cid,
				}

				rw.WriteHeader(http.StatusOK)
				ceh.ProblemList.Execute(rw, templateVal)

				return
			}
			pidx, err := strconv.ParseInt(req.URL.Path, 10, 64)

			if err != nil {
				sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

				return
			}

			prob, err := mainDB.ContestProblemFind2(cid, pidx)

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
				ContestProblem
				ContestName string
				Cid         int64
				Text        string
				UserName    string
			}
			templateVal := TemplateVal{*prob, cont.Name, cid, string(html), std.UserName}

			rw.WriteHeader(http.StatusOK)

			ceh.ProblemView.Execute(rw, templateVal)
		})).ServeHTTP(rw, req)
	})

	mux.Handle("/ranking", http.StripPrefix("/ranking", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !free {
			RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/")

			return
		}

		wrapForm := createWrapForm(req)

		page := int(wrapForm("p"))

		if page == -1 {
			page = 1
		}

		type RankingRow2 struct {
			RankingRow
			Rank int
		}

		type TemplateVal struct {
			ContestName string
			Cid         int64
			UserName    string
			Problems    []ContestProblem
			Ranking     []RankingRow2
			Current     int
			MaxPage     int
			BeginTime   int64
			Pagination  []PaginationHelper
		}

		count, err := mainDB.ContestRankingCount(cid)

		if err != nil {
			DBLog().WithError(err).WithField("iid", std.Iid).Error("ContestRankingCount error")
			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

			return
		}

		var templateVal TemplateVal
		templateVal.Cid = cid
		templateVal.ContestName = cont.Name
		templateVal.UserName = std.UserName
		templateVal.BeginTime = cont.StartTime.Unix()
		probs, err := mainDB.ContestProblemList(cid)

		if err != nil {
			DBLog().WithError(err).WithField("iid", std.Iid).Error("ContestProblemList error")
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
			if (page-1)*ContentsPerPage > int(count) {
				page = 1
			}

			templateVal.Current = page

			ranks, err := mainDB.ContestRankingList(cid, int64((page-1)*ContentsPerPage), ContentsPerPage)

			if err != nil {
				DBLog().WithError(err).Error("ContestRankingList error")

				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			ranks2 := make([]RankingRow2, len(ranks))

			for i := range ranks {
				ranks2[i] = RankingRow2{ranks[i], (page-1)*ContentsPerPage + i + 1}
			}

			templateVal.Ranking = ranks2
		}

		templateVal.Pagination = NewPaginationHelper(templateVal.Current, templateVal.MaxPage, 3)

		ceh.RankingPage.Execute(rw, templateVal)
	})))

	mux.Handle("/submissions/", http.StripPrefix("/submissions/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !free {
			RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/")

			return
		}

		if req.URL.Path == "" {
			wrapForm := createWrapForm(req)

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

			if !(isFinished || isAdmin) && iid != std.Iid {
				RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/submissions/?user="+std.UserID)

				return
			}

			count, err := mainDB.SubmissionViewCount(cid, iid, lang, prob, stat)

			if err != nil {
				DBLog().WithError(err).WithField("iid", std.Iid).Error("SubmissionViewCount error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			type TemplateVal struct {
				AllEnabled  bool
				ContestName string
				UserName    string
				Cid         int64
				Uid         string
				Submissions []SubmissionView
				Problems    []ContestProblem
				Languages   []Language
				Current     int
				MaxPage     int
				Pagination  []PaginationHelper
				Lang        int64
				Prob        int64
				Status      int64
				User        string
			}
			var templateVal TemplateVal
			templateVal.AllEnabled = isFinished || isAdmin
			templateVal.ContestName = cont.Name
			templateVal.Cid = cid
			templateVal.UserName = std.UserName
			templateVal.User = userID
			templateVal.Status = stat
			templateVal.Lang = lang
			templateVal.Prob = prob
			templateVal.Uid = std.UserID

			langs, err := mainDB.LanguageActiveList()

			if err != nil {
				HttpLog().WithField("iid", iid).WithError(err).Error("LanguageList() error")
			} else {
				templateVal.Languages = langs
			}

			probs, err := mainDB.ContestProblemListLight(cid)

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

				submissions, err := mainDB.SubmissionViewList(cid, iid, lang, prob, stat, int64((page-1)*ContentsPerPage), ContentsPerPage)

				if err == nil {
					templateVal.Submissions = submissions
				} else {
					HttpLog().WithField("iid", iid).WithError(err).Error("SubmissionViewList() error")
				}
			}

			templateVal.Pagination = NewPaginationHelper(templateVal.Current, templateVal.MaxPage, 3)

			rw.WriteHeader(200)

			ceh.SubmissionList.Execute(rw, templateVal)
		} else {
			sid, err := strconv.ParseInt(req.URL.Path, 10, 64)

			if err != nil {
				sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

				return
			}

			submission, err := mainDB.SubmissionViewFind(sid, cid)

			if err == ErrUnknownSubmission {
				sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

				return
			} else if err != nil {
				DBLog().WithError(err).WithField("sid", sid).Error("SubmissionViewFind error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}

			if submission.Cid != cid {
				sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

				return
			}

			if !isAdmin && submission.Iid != std.Iid && !isFinished {
				sctypes.ResponseTemplateWrite(http.StatusForbidden, rw)

				return
			}

			code, err := mainDB.SubmissionGetCode(sid)

			if err != nil {
				var tmp string

				code = tmp
			}

			type SubmissionTestCaseView struct {
				SubmissionTestCase
				StatusString string
			}

			casesArr, err := mainDB.SubmissionGetCase(sid)
			var caseViews []SubmissionTestCaseView

			if err == nil {
				caseViews = make([]SubmissionTestCaseView, 0, len(casesArr))
				for _, v := range casesArr {
					caseViews = append(caseViews, SubmissionTestCaseView{v, sctypes.SubmissionStatusTypeToString[v.Status]})
				}
			} else {
				HttpLog().WithError(err).Error("SubmissionGetCase() error")
			}

			msg, err := mainDB.SubmissionGetMsg(sid)

			if err != nil {
				msg = ""
			}

			type TemplateVal struct {
				ContestName string
				Submission  SubmissionView
				Cases       []SubmissionTestCaseView
				Code        string
				Msg         string
				UserName    string
				Cid         int64
			}

			templateVal := TemplateVal{
				ContestName: cont.Name,
				Submission:  *submission,
				Cases:       caseViews,
				Code:        code,
				Msg:         msg,
				UserName:    std.UserName,
				Cid:         cid,
			}

			rw.WriteHeader(http.StatusOK)
			ceh.SubmissionView.Execute(rw, templateVal)
		}
	})))

	mux.HandleFunc("/join", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" && !isAdmin {
			mainDB.ContestParticipationAdd(std.Iid, cid)

			RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/")
		} else {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}
	})

	mux.HandleFunc("/submit", func(rw http.ResponseWriter, req *http.Request) {
		if !free {
			RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/")

			return
		}

		if req.Method == "GET" {
			type TemplateVal struct {
				ContestName string
				UserName    string
				Cid         int64
				Problems    []ContestProblem
				Languages   []Language
				Prob        int64
			}

			list, err := mainDB.ContestProblemListLight(cid)

			if err != nil {
				list = []ContestProblem{}

				HttpLog().WithError(err).Error("ContestProblemListLight() error")
			}

			lang, err := mainDB.LanguageActiveList()

			if err != nil {
				lang = []Language{}

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
				cont.Name,
				std.UserName,
				cid,
				list,
				lang,
				prob,
			}

			rw.WriteHeader(http.StatusOK)
			ceh.SubmitPage.Execute(rw, templateVal)
		} else if req.Method == "POST" {
			wrapForm := createWrapForm(req)

			wrapFormStr := createWrapFormStr(req)

			lid := wrapForm("lang")
			pid := wrapForm("prob")
			code := wrapFormStr("code")

			if lid < 0 || pid < 0 || code == "" {
				sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

				return
			}

			prob, err := mainDB.ContestProblemFind2(cid, pid)

			if err != nil {
				if err == ErrUnknownProblem {
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
				if err == ErrUnknownLanguage {
					rw.WriteHeader(http.StatusBadRequest)
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				} else {
					DBLog().WithError(err).Error("LanguageFind error")
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					return
				}
			}

			subm, err := mainDB.SubmissionAdd(prob.Pid, std.Iid, lid, code)

			if err != nil {
				DBLog().WithError(err).Error("SubmissionAdd error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return
			}
			//SJQueue.Push(subm)

			RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/submissions/"+strconv.FormatInt(subm, 10))
		} else {

		}
	})

	mux.Handle("/management/", http.StripPrefix("/management/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isAdmin {
			sctypes.ResponseTemplateWrite(http.StatusForbidden, rw)

			return
		}

		if req.URL.Path == "" {
			type TemplateVal struct {
				Cid         int64
				UserName    string
				ContestName string
			}
			ceh.ManagementTopPage.Execute(rw, TemplateVal{cid, std.UserName, cont.Name})
		} else if req.URL.Path == "remove" {
			list, err := mainDB.ContestProblemList(cid)

			if err != nil {
				DBLog().WithError(err).Error("ContestProblemList() error")
			}

			for i := range list {
				err = mainDB.SubmissionRemoveAll((list)[i].Pid)

				if err != nil {
					DBLog().WithError(err).Error("SubmissionRemoveAll() error")
				}
			}

			err = mainDB.ContestParticipationRemove(cid)

			if err != nil {
				DBLog().WithError(err).Error("ContestParticipationRemove() error")
			}

			err = mainDB.ContestDelete(cid)

			if err != nil {
				DBLog().WithError(err).Error("ContestDelete() error")
			}

			RespondRedirection(rw, "/contests/")

			return
		} else if req.URL.Path == "rejudge" {
			respondTemp := func(msg string) {
				type TemplateVal struct {
					Cid         int64
					UserName    string
					Msg         *string
					ContestName string
				}

				if msg == "" {
					ceh.ManagementRejudgePage.Execute(rw, TemplateVal{cid, std.UserName, nil, cont.Name})
				} else {
					ceh.ManagementRejudgePage.Execute(rw, TemplateVal{cid, std.UserName, &msg, cont.Name})
				}
			}

			if req.Method == "GET" {
				respondTemp("")

				return
			} else if req.Method == "POST" {
				wrapForm := createWrapForm(req)

				target, id := wrapForm("target"), wrapForm("id")

				if (target != 1 && target != 2) || id < 0 {
					respondTemp("不正なIDです。")

					return
				}

				if target == 1 {
					sm, err := mainDB.SubmissionFind(id)

					if err != nil {
						if err == ErrUnknownSubmission {
							respondTemp("該当する提出がありません。")
						} else {
							DBLog().WithError(err).Error("SubmissionFind error")
							sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)
						}
						return
					}

					cp, err := mainDB.ContestProblemFind(sm.Pid)

					if err != nil {
						DBLog().WithError(err).Error("ContestProblemFind error")
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}

					if cp.Cid != cid {
						respondTemp("該当する提出がありません。")

						return
					}

					//SJQueue.Push(sm.Sid)

					RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/management/")

					return
				} else {
					cp, err := mainDB.ContestProblemFind2(cid, id)

					if err != nil {
						if err == ErrUnknownProblem {
							respondTemp("該当する問題がありません。")
						} else {
							DBLog().WithError(err).Error("ContestProblemFind2 error")
							sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						}
						return
					}

					sml, err := mainDB.SubmissionListWithPid(cp.Pid)

					if err != nil {
						DBLog().WithError(err).Error("SubmissionList error")
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}

					for _ = range *sml {
						// TODO: JudgeQueue処理
						//SJQueue.Push((*sml)[i].Sid)
					}

					RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/management/")

					return
				}

				RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/management/")

			} else {
				sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

				return
			}
		} else if req.URL.Path == "setting" {
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
						cid, std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManagementSettingPage.Execute(rw, templateVal)

					return
				}

				start, err := time.ParseInLocation("2006/01/02 15:04", startStr, Location)

				if err != nil {
					msg := "開始日時の値が不正です。"
					templateVal := TemplateVal{
						cid, std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManagementSettingPage.Execute(rw, templateVal)

					return
				}

				if cont.StartTime.Unix() <= time.Now().Add(2*time.Minute).Unix() && cont.StartTime.Unix() != start.Unix() {
					msg := "開始日時は2分前を切ると変更できません。"

					startDate = cont.StartTime.In(Location).Format("2006/01/02")
					startTime = cont.StartTime.In(Location).Format("15:04")

					templateVal := TemplateVal{
						cid, std.UserID, &msg, startDate, cont.StartTime.In(Location).Format("2006/01/02 15:04"), finishDate, finishTime, description, contestName,
					}

					ceh.ManagementSettingPage.Execute(rw, templateVal)

					return
				}

				finish, err := time.ParseInLocation("2006/01/02 15:04", finishStr, Location)

				if err != nil {
					msg := "終了日時の値が不正です。"
					templateVal := TemplateVal{
						cid, std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManagementSettingPage.Execute(rw, templateVal)

					return
				}

				if cont.FinishTime.Unix() <= time.Now().Add(2*time.Minute).Unix() && cont.FinishTime.Unix() != finish.Unix() {
					msg := "終了日時は2分前を切ると変更できません。"

					finishDate = cont.FinishTime.In(Location).Format("2006/01/02")
					finishTime = cont.FinishTime.In(Location).Format("15:04")

					templateVal := TemplateVal{
						cid, std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManagementSettingPage.Execute(rw, templateVal)

					return
				}

				if start.Unix() >= finish.Unix() || (cont.StartTime.Unix() != start.Unix() && start.Unix() < time.Now().Unix()) || (cont.FinishTime.Unix() != finish.Unix() && finish.Unix() < time.Now().Unix()) {
					msg := "開始日時及び終了日時の値が不正です。"
					templateVal := TemplateVal{
						cid, std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManagementSettingPage.Execute(rw, templateVal)

					return
				}

				err = mainDB.ContestUpdate(cid, contestName, start, finish, cont.Admin, 0)

				if err != nil {
					if strings.Index(err.Error(), "Duplicate") != -1 {
						msg := "すでに存在するコンテスト名です。"
						templateVal := TemplateVal{
							cid, std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
						}

						ceh.ManagementSettingPage.Execute(rw, templateVal)

						return
					} else {
						DBLog().WithError(err).Error("ContestUpdate error")
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}
				}

				err = (&Contest{Cid: cid}).DescriptionUpdate(description)

				if err != nil {
					HttpLog().WithError(err).Error("DescriptionUpdate() error")
				}

				RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/management/")
			} else if req.Method == "GET" {
				desc, _ := (&Contest{Cid: cid}).DescriptionLoad()

				templateVal := TemplateVal{
					Cid:         cid,
					UserName:    std.UserID,
					StartDate:   cont.StartTime.In(Location).Format("2006/01/02"),
					StartTime:   cont.StartTime.In(Location).Format("15:04"),
					FinishDate:  cont.FinishTime.In(Location).Format("2006/01/02"),
					FinishTime:  cont.FinishTime.In(Location).Format("15:04"),
					ContestName: cont.Name,
					Description: desc,
				}
				ceh.ManagementSettingPage.Execute(rw, templateVal)
			} else {
				sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

				return
			}
		} else if len(req.URL.Path) >= 9 && req.URL.Path[:9] == "problems/" {
			http.StripPrefix("problems/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "" {
					type TemplateVal struct {
						Cid         int64
						ContestName string
						UserName    string
						Problems    []ContestProblem
					}

					list, err := mainDB.ContestProblemList(cid)

					if err != nil {
						DBLog().WithError(err).Error("ContestProblemList error")
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}

					ceh.ManagementProblemList.Execute(rw, TemplateVal{cid, cont.Name, std.UserName, list})
				} else if upidx, err := strconv.ParseInt(req.URL.Path, 10, 64); req.URL.Path == "new" || err == nil {
					if err != nil {
						upidx = -1

						cnt, err := mainDB.ContestProblemCount(cid)

						if err != nil {
							DBLog().WithError(err).Error("ContestProblemCoun error")
							sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

							return
						}

						if cnt >= 50 {
							sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

							return
						}
					}

					type TemplateVal struct {
						Cid         int64
						ContestName string
						UserName    string
						Msg         *string
						Mode        bool
						Pidx        int64
						Name        string
						Time        int64
						Mem         int64
						Type        int64
						Prob        string
						Lang        int64
						Languages   []Language
						Code        string
					}

					wrapForm := createWrapForm(req)

					wrapFormStr := createWrapFormStr(req)

					languages, err := mainDB.LanguageList()

					if err != nil {
						DBLog().WithError(err).Error("LanguageList error")
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}

					var cp *ContestProblem
					if upidx != -1 {
						cp, err = mainDB.ContestProblemFind2(cid, upidx)

						if err != nil {
							if err == ErrUnknownProblem {
								sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

								return
							} else {
								DBLog().WithError(err).Error("ContestProblemFind2 error")
								sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

								return
							}
						}
					}

					if req.Method == "GET" {
						temp := TemplateVal{Cid: cid, ContestName: cont.Name, Time: 1, Mem: 32, UserName: std.UserName, Mode: true, Languages: languages}

						if upidx != -1 {
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
							temp.Mode = false
							temp.Name = cp.Name
							temp.Time = cp.Time
							temp.Mem = cp.Mem
							temp.Pidx = upidx
							temp.Type = int64(cp.Type)
							temp.Lang = lid
							temp.Code = checker
							temp.Prob = stat

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
							msg := "問題名が不正です。"
							mode := false
							if upidx == -1 {
								mode = true
							}
							ceh.ManagementProblemSettingPage.Execute(rw, TemplateVal{cid, cont.Name, std.UserName, &msg, mode, pidx, name, time, mem, jtype, prob, lid, languages, code})

							return
						}

						if sctypes.JudgeType(jtype) == sctypes.JudgeRunningCode {
							if _, err := mainDB.LanguageFind(lid); err != nil {
								if err == ErrUnknownLanguage {
									sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

									return
								} else {
									DBLog().WithError(err).Error("LanguageFind error")
									sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

									return
								}
							}
						}

						if upidx != -1 {
							cp.Pidx = pidx
							cp.Name = name
							cp.Time = time
							cp.Mem = mem
							cp.Type = sctypes.JudgeType(jtype)

							err = mainDB.ContestProblemUpdate(*cp)
						} else {
							cp, err = cont.ProblemAdd(pidx, name, time, mem, sctypes.JudgeType(jtype))
						}

						if err != nil {
							if strings.Index(err.Error(), "Duplicate") != -1 {
								msg := "使用されている問題番号です。"
								mode := false
								if upidx == -1 {
									mode = true
								}
								ceh.ManagementProblemSettingPage.Execute(rw, TemplateVal{cid, cont.Name, std.UserName, &msg, mode, pidx, name, time, mem, jtype, prob, lid, languages, code})

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

						RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/management/problems/")
					} else {
						sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

						return
					}
				}
			})).ServeHTTP(rw, req)
		} else if len(req.URL.Path) >= 11 && req.URL.Path[:10] == "testcases/" {
			arr := strings.Split(req.URL.Path[10:], "/")

			pidx, err := strconv.ParseInt(arr[0], 10, 64)

			if err != nil {
				sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

				return
			}

			cp, err := mainDB.ContestProblemFind2(cid, pidx)

			if err == ErrUnknownProblem {
				sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

				return
			} else if err != nil {
				DBLog().WithError(err).Error("ContestProblemFind2 error")
				sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

				return

			}

			if len(arr) == 1 {
				type TemplateVal struct {
					Cid         int64
					Pidx        int64
					ContestName string
					UserName    string
					Testcases   []string
					Scoresets   []ContestProblemScoreSet
					Msg         *string
				}

				if req.Method == "GET" {
					cases, sets, err := cp.LoadTestCaseNames()

					if err != nil {
						DBLog().WithError(err).Error("LoadTestCaseNames error")
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}

					ceh.ManagementTastcaseList.Execute(rw, TemplateVal{cid, pidx, cont.Name, std.UserName, cases, sets, nil})
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

					scores := make([]ContestProblemScoreSet, len(setScores))
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

						scores[i] = ContestProblemScoreSet{
							Score: score,
						}

						scores[i].Cases.Set(caseIds)
						scores[i].BeforeSave() // copy from cases to casesrawstring
					}

					if illegal {
						msg := "不正なパラメータがあります。"

						ceh.ManagementTastcaseList.Execute(rw, TemplateVal{cid, pidx, cont.Name, std.UserName, cases, scores, &msg})

						return
					}

					err := cp.UpdateTestCaseNames(cases, scores)

					if err != nil {
						DBLog().WithError(err).Error("UpdateTestCaseNames error")
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}

					RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/management/testcases/"+strconv.FormatInt(pidx, 10))
				} else {
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				}
			} else if len(arr) >= 2 {
				tcid, err := strconv.ParseInt(arr[1], 10, 64)

				if err != nil {
					sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

					return
				}

				if len(arr) == 2 {

					in, out, err := cp.LoadTestCaseInfo(int(tcid))

					if err == ErrUnknownTestcase {
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
						cid,
						int(tcid),
						std.UserName,
						pidx,
						cp.Name,
						in, out,
					}

					ceh.ManagementTestcaseSetting.Execute(rw, templateVal)
				} else if len(arr) == 3 {
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
						} else if l > 20*1024*1024 {
							sctypes.ResponseTemplateWrite(http.StatusRequestEntityTooLarge, rw)

							return
						}

						file.Seek(0, 0)

						defer file.Close()

						if arr[2] == "input" {
							err = cp.UpdateTestCase(true, tcid, NewTrimNewlineReader(file))
						} else if arr[2] == "output" {
							err = cp.UpdateTestCase(false, tcid, NewTrimNewlineReader(file))
						} else {
							sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

							return
						}

						if err == ErrUnknownTestcase {
							sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

							return
						} else if err != nil {
							DBLog().WithError(err).Error("UpdateTestCase error")
							sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

							return
						}

						RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/management/testcases/"+strconv.FormatInt(pidx, 10)+"/"+strconv.FormatInt(int64(tcid), 10))
					} else if req.Method == "GET" {
						var reader io.ReadCloser
						var err error

						if arr[2] == "input" {
							reader, err = cp.LoadTestCase(true, int(tcid))
						} else if arr[2] == "output" {
							reader, err = cp.LoadTestCase(false, int(tcid))
						} else {
							sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

							return
						}

						if err == ErrUnknownTestcase {
							sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

							return
						} else if err != nil {
							DBLog().WithError(err).Error("UpdateTestCase error")
							sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

							return
						}
						defer reader.Close()

						fileName := strconv.FormatInt(cid, 10) + "-" + strconv.FormatInt(pidx, 10) + "_" + strconv.FormatInt(int64(tcid), 10)

						if arr[2] == "input" {
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
						sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

						return
					}
				} else {
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				}
			}

		} else {
			sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

			return
		}
	})))

	handler := func(rw http.ResponseWriter, req *http.Request) {
		mux.ServeHTTP(rw, req)
	}

	return handler, nil
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

	return &ContestEachHandler{top.Lookup("index_tmpl.html"), probList, probView, subList.Lookup("submissions_tmpl.html"), subView.Lookup("submission_view_tmpl.html"), submit, man, manre, manse.Lookup("setting_tmpl.html"), manpr, manprv, mantc, mantcv, rank.Lookup("ranking_tmpl.html")}, nil
}
