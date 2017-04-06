package main

import (
	"net/http"
	"text/template"

	"context"

	"strings"

	"github.com/cs3238-tsuzu/popcon-sc/setting"
	"github.com/cs3238-tsuzu/popcon-sc/types"
	muxlib "github.com/gorilla/mux"
)

func AdminHandler() (http.Handler, error) {
	mux := muxlib.NewRouter()
	handler := func(rw http.ResponseWriter, req *http.Request) {
		std, err := ParseRequestForSession(req)

		if err == ErrUnknownSession {
			RespondRedirection(rw, "/login?comeback=/admin/")

			return
		}

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}

		if std.Gid != 0 {
			sctypes.ResponseTemplateWrite(http.StatusForbidden, rw)

			return
		}
		mux.ServeHTTP(rw, req.WithContext(context.WithValue(context.Background(), ContextValueKeySessionTemplateData, *std)))
	}

	sessionTemplateData := func(req *http.Request) *SessionTemplateData {
		v := req.Context().Value(ContextValueKeySessionTemplateData)

		if std, ok := v.(SessionTemplateData); ok {
			return &std
		}

		return nil
	}

	err := ProcessUntilError(
		func() error {
			tmpl, err := template.ParseFiles("./html/admin/index_tmpl.html")

			if err != nil {
				return err
			}

			type TemplateVal struct {
				UserName string
			}

			mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
				// Mustn't be nil
				std := sessionTemplateData(req)

				tmpl.Execute(rw, TemplateVal{
					UserName: std.UserName,
				})
			})

			return nil
		},
		func() error {
			tmpl, err := template.ParseFiles("./html/admin/general_tmpl.html")

			if err != nil {
				return err
			}

			type TemplateVal struct {
				Error    string
				UserName string
				Setting  *ppconfiguration.Structure
				Groups   []Group
			}

			mux.HandleFunc("/general", func(rw http.ResponseWriter, req *http.Request) {
				std := sessionTemplateData(req)

				var val TemplateVal
				val.UserName = std.UserName
				if req.Method == "GET" {
					val.Setting, err = mainRM.GetAll()

					if err != nil {
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)
						DBLog().WithError(err).Error("GetAll() error")

						return
					}

					val.Groups, err = mainDB.GroupList()
					if err != nil {
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)
						DBLog().WithError(err).Error("GroupList() error")

						return
					}

					tmpl.Execute(rw, val)
				} else if req.Method == "POST" {
					if err := req.ParseForm(); err != nil {
						sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)
						return
					}
					parseForm := createWrapFormInt(req)
					parseFormInt64 := createWrapFormInt64(req)
					parseFormStr := createWrapFormStr(req)

					canCreateUser := parseForm(string(ppconfiguration.CanCreateUser)) == 1
					canCreateContest := parseForm(string(ppconfiguration.CanCreateContest)) == 1
					numberOfDisplayedNews := parseForm(string(ppconfiguration.NumberOfDisplayedNews))
					certificationWithEmail := parseForm(string(ppconfiguration.CertificationWithEmail)) == 1
					sendMailCommand := strings.Split(parseFormStr(string(ppconfiguration.SendMailCommand)), ",")
					csrfConfTokenExpiration := parseFormInt64(string(ppconfiguration.CSRFConfTokenExpiration))
					mailConfTokenExpiration := parseFormInt64(string(ppconfiguration.MailConfTokenExpiration))
					mailMinInterval := parseFormInt64(string(ppconfiguration.MailMinInterval))
					sessionExpiration := parseForm(string(ppconfiguration.SessionExpiration))
					standardSignupGroup := parseFormInt64(string(ppconfiguration.StandardSignupGroup))
					publicHost := parseFormStr(string(ppconfiguration.PublicHost))

					var val TemplateVal

					failureCheck := func() string {
						var errs []string
						val.Setting.CanCreateUser = canCreateUser
						val.Setting.CanCreateContest = canCreateContest
						val.Setting.CertificationWithEmail = certificationWithEmail
						val.Setting.SendMailCommand = sendMailCommand
						val.Setting.PublicHost = publicHost

						if numberOfDisplayedNews > 0 && numberOfDisplayedNews <= 1000 {
							val.Setting.NumberOfDisplayedNews = numberOfDisplayedNews
						} else {
							errs = append(errs, "ニュース表示数の値が不正です。")
						}
						if csrfConfTokenExpiration > 0 {
							val.Setting.CSRFConfTokenExpiration = csrfConfTokenExpiration
						} else {
							errs = append(errs, "CSRFトークン有効期限の値が不正です。")
						}
						if mailConfTokenExpiration > 0 {
							val.Setting.MailConfTokenExpiration = mailConfTokenExpiration
						} else {
							errs = append(errs, "メール認証トークン有効期限の値が不正です。")
						}
						if mailMinInterval > 0 {
							val.Setting.MailMinInterval = mailMinInterval
						} else {
							errs = append(errs, "メール送信最小インターバルの値が不正です。")
						}

						if sessionExpiration > 0 {
							val.Setting.SessionExpiration = sessionExpiration
						} else {
							errs = append(errs, "セッション有効期限の値が不正です。")
						}

						if standardSignupGroup > 0 {
							_, err := mainDB.GroupFind(standardSignupGroup)

							if err == ErrUnknownGroup {
								errs = append(errs, "登録されていない、または削除されたグループです。")
							} else {
								if err != nil {
									DBLog().WithError(err).Error("GroupFind() error")
								}
								val.Setting.StandardSignupGroup = standardSignupGroup
							}
						}

						return strings.Join(errs, "<br>")
					}

					if msg := failureCheck(); len(msg) == 0 {
						err := mainRM.SetAll(val.Setting)

						if err != nil {
							DBLog().WithError(err).Error("SetAll() error")

							sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

							return
						}

						RespondRedirection(rw, "/admin/")
					} else {

						val.Groups, err = mainDB.GroupList()
						if err != nil {
							sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)
							DBLog().WithError(err).Error("GroupList() error")

							return
						}

						tmpl.Execute(rw, val)
					}
				} else {
					sctypes.ResponseTemplateWrite(http.StatusNotImplemented, rw)

					return
				}
			})

			return nil
		})

	if err != nil {
		return nil, err
	}
	return http.HandlerFunc(handler), nil
}
