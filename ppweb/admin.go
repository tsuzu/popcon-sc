package main

import (
	"net/http"
	"strconv"
	"text/template"

	"context"

	"strings"

	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
	"github.com/cs3238-tsuzu/popcon-sc/lib/setting"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/gorilla/mux"
)

func AdminHandler() (http.Handler, error) {
	router := mux.NewRouter()
	handler := func(rw http.ResponseWriter, req *http.Request) {
		std, err := ParseRequestForSession(req)

		if err == database.ErrUnknownSession {
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
		router.ServeHTTP(rw, req.WithContext(context.WithValue(context.Background(), ContextValueKeySessionTemplateData, *std)))
	}

	sessionTemplateData := func(req *http.Request) *database.SessionTemplateData {
		v := req.Context().Value(ContextValueKeySessionTemplateData)

		if std, ok := v.(database.SessionTemplateData); ok {
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

			router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
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
				Groups   []database.Group
			}

			router.HandleFunc("/general", func(rw http.ResponseWriter, req *http.Request) {
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

					val.Setting = &ppconfiguration.Structure{}

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

							if err == database.ErrUnknownGroup {
								errs = append(errs, "登録されていない、または削除されたグループです。")
							} else {
								if err != nil {
									DBLog().WithError(err).Error("GroupFind() error")
								}
								val.Setting.StandardSignupGroup = standardSignupGroup
							}
						} else {
							errs = append(errs, "不正なグループが指定されています。")
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
						val.Error = msg

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
		},
		func() error {
			/*tmpl, err := template.ParseFiles("./html/admin/users_tmpl.html")

			if err != nil {
				return err
			}*/

			/*router.HandleFunc("/users/", func(rw http.ResponseWriter, req *http.Request) {
				if err := req.ParseForm(); err != nil {
					sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

					return
				}

				wrapFormInt64 := createWrapFormInt64(req)

				page := wrapFormInt64("page")

				if page <= 0 {
					page = 1
				}

				if err := mainDB.UserCount(); err != nil {
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					DBLog().WithError(err).Error("UserCount() error")
				}

				users, err := mainDB.UserList(ContentsPerPage, ContentsPerPage * (page - 1))

				if err != nil {
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					DBLog().WithError(err).Error("UserList() error")
				}
			})*/

			return nil
		},
		func() error {
			tmpl, err := template.ParseFiles("./html/admin/languages_tmpl.html")

			if err != nil {
				return err
			}

			type TemplateVal struct {
				UserName  string
				Languages []database.Language
			}

			router.HandleFunc("/languages/", func(rw http.ResponseWriter, req *http.Request) {
				languages, err := mainDB.LanguageList()

				if err != nil {
					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					DBLog().WithError(err).Error("LanguageList() error")

					return
				}

				std := sessionTemplateData(req)

				tmpl.Execute(rw, TemplateVal{
					UserName:  std.UserName,
					Languages: languages,
				})
			})

			tmpl2, err := template.ParseFiles("./html/admin/language_setting_tmpl.html")

			if err != nil {
				return err
			}

			type TemplateValSetting struct {
				IsAddition bool
				UserName   string
				Msg        string
				Language   database.Language
			}

			router.HandleFunc("/languages/new", func(rw http.ResponseWriter, req *http.Request) {
				std := sessionTemplateData(req)

				if req.Method == "GET" {
					tmpl2.Execute(rw, TemplateValSetting{
						IsAddition: true,
						UserName:   std.UserName,
					})

					return
				}

				if req.Method == "POST" {
					if err := req.ParseForm(); err != nil {
						sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

						return
					}

					wrapForm := createWrapFormStr(req)

					lang := wrapForm("language_name")
					highlightType := wrapForm("highlight_type_name")
					active := wrapForm("active")

					_, err := mainDB.LanguageAdd(lang, highlightType, active != "0")

					if err != nil {
						tmpl2.Execute(rw, TemplateValSetting{
							UserName: std.UserName,
							Language: database.Language{
								Name:          lang,
								HighlightType: highlightType,
								Active:        active != "0",
							},
							Msg: "エラーが発生しました。",
						})
					}

					RespondRedirection(rw, "/admin/languages")
				}
			})

			router.HandleFunc("/languages/{lid:[0-9]+}", func(rw http.ResponseWriter, req *http.Request) {
				lid, _ := strconv.ParseInt(mux.Vars(req)["lid"], 10, 64)
				std := sessionTemplateData(req)

				lang, err := mainDB.LanguageFind(lid)

				if err != nil {
					if err == database.ErrUnknownLanguage {
						sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)

						return
					}

					sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

					DBLog().WithError(err).Error("LanguageFind() error")

					return
				}

				templateVal := TemplateValSetting{
					IsAddition: false,
					UserName:   std.UserName,
					Language:   *lang,
				}

				if req.Method == "GET" {
					tmpl2.Execute(rw, templateVal)
				} else if req.Method == "POST" {
					if err := req.ParseForm(); err != nil {
						sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

						return
					}
					wrapForm := createWrapFormStr(req)

					lang := wrapForm("language_name")
					highlightType := wrapForm("highlight_type_name")
					active := wrapForm("active")

					if err := mainDB.LanguageUpdate(lid, lang, highlightType, active != "0"); err != nil {
						sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

						return
					}

					RespondRedirection(rw, "/admin/languages/")
				}
			})

			return nil
		},
	)

	if err != nil {
		return nil, err
	}
	return http.HandlerFunc(handler), nil
}
