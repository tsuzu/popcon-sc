package main

import (
	"bytes"
	"crypto/sha512"
	"errors"
	"fmt"
	html "html/template"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	"github.com/Sirupsen/logrus"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

var UTF8BOM = []byte{239, 187, 191}
var UTF8EDL = []byte{'\r', '\n'}

func HasBOM(in []byte) bool {
	return bytes.HasPrefix(in, UTF8BOM)
}

func StripBOM(in []byte) []byte {
	return bytes.TrimPrefix(in, UTF8BOM)
}

func StripEndline(in []byte) []byte {
	return bytes.TrimPrefix(in, UTF8EDL)
}

func UTF8StringLengthAndBOMCheck(str string, l int) bool {
	if len(str) > l*6 {
		return false
	}

	if utf8.RuneCountInString(str) > l {
		return false
	}

	return !HasBOM([]byte(str))
}

func ReplaceEndline(str string) string {
	return strings.Replace(strings.Replace(str, "\r\n", "\n", -1), "\r", "\n", -1)
}

func TimeToString(t int64) string {
	return time.Unix(t, 0).In(Location).Format("2006/01/02 15:04:05")
}

// CreateHandlers is a function to return hadlers
func CreateHandlers() (map[string]http.Handler, error) {
	res := make(map[string]http.Handler)

	var err error

	res["/"], err = func() (http.Handler, error) {
		funcs := template.FuncMap{
			"timeToString": TimeToString,
		}

		temp, err := template.New("").Funcs(funcs).ParseFiles("./html/index_tmpl.html")

		if err != nil {
			return nil, err
		}

		tmp := temp.Lookup("index_tmpl.html")

		if err != nil {
			return nil, errors.New("Failed to load ./html/index_tmpl.html")
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.Path != "/" && req.URL.Path != "/#" {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(NF404))

				return
			}

			std, err := ParseRequestForSession(req)

			if std == nil || err != nil {
				std = &SessionTemplateData{
					IsSignedIn: false,
					UserID:     "",
					UserName:   "",
				}
			}

			cnt := settingManager.Get().NumberOfDisplayedNews

			news, err := mainDB.NewsGet(cnt)

			if err != nil {
				news = make([]News, 0)

				DBLog.WithError(err).Error("NewsGet failed")
			}

			type IndexResp struct {
				*SessionTemplateData
				News      []News
				NewsCount int
			}

			resp := &IndexResp{
				SessionTemplateData: std,
				News:                news,
				NewsCount:           cnt,
			}

			rw.WriteHeader(http.StatusOK)
			tmp.Execute(rw, *resp)
		})

		return f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/onlinejudge/"], err = func() (http.Handler, error) {
		ojh := http.StripPrefix("/onlinejudge/", CreateOnlineJudgeHandler())

		if ojh == nil {
			return nil, errors.New("Failed to CreateOnlineJudgeHandler()")
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusNotImplemented)

			fmt.Fprint(rw, NI501)

			return

			ojh.ServeHTTP(rw, req)
		})

		return f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/contests/"], err = func() (http.Handler, error) {
		contestsTopHandler, err := CreateContestsTopHandler()

		if err != nil {
			return nil, err
		}

		return http.StripPrefix("/contests", *contestsTopHandler), nil
	}()

	if err != nil {
		return nil, err
	}

	res["/login"], err = func() (http.HandlerFunc, error) {
		type LoginTemp struct {
			BackURL       string
			Error         string
			SignupEnabled bool
		}

		tmp, err := template.ParseFiles("./html/login_tmpl.html")

		if err != nil {
			return nil, errors.New("Failed to load ./html/login_tmpl.html")
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Method == "GET" {
				req.ParseForm()
				rw.WriteHeader(http.StatusOK)

				comeback, res := req.Form["comeback"]

				var cburl string
				if !res || len(comeback) == 0 || len(comeback[0]) == 0 {
					cburl = "/"
				} else {
					cburl = comeback[0]
				}

				tmp.Execute(rw, LoginTemp{
					BackURL:       cburl,
					SignupEnabled: settingManager.Get().CanCreateUser,
				})
			} else if req.Method == "POST" {
				if err := req.ParseForm(); err != nil {
					rw.WriteHeader(http.StatusBadRequest)

					return
				}

				loginID := req.FormValue("loginID")
				password := req.FormValue("password")
				comeback := req.FormValue("comeback")

				if len(comeback) == 0 {
					comeback = "/"
				}

				if strings.Index(comeback, "//") != -1 || len(comeback) > 128 {
					rw.WriteHeader(http.StatusBadRequest)
					fmt.Fprint(rw, BR400)

					return
				}

				user, err := mainDB.UserFindFromUserID(loginID)
				passHash := sha512.Sum512([]byte(password))

				if err != nil || !reflect.DeepEqual(user.PassHash, passHash[:]) {
					rw.WriteHeader(http.StatusOK)

					tmp.Execute(rw, LoginTemp{
						Error:         "IDまたはパスワードが間違っています。",
						BackURL:       comeback,
						SignupEnabled: settingManager.Get().CanCreateUser,
					})
					return
				}

				if !user.Enabled {
					rw.WriteHeader(http.StatusOK)

					var msg string
					if user.Email.Valid {
						err := MailSendConfirmUser(user.Iid, user.UserName, user.Email.String)

						if err != nil {
							if err != ErrMailWasSent {
								rw.WriteHeader(http.StatusInternalServerError)
								rw.Write([]byte(ISE500))
								return
							}

							msg = "アカウントが有効化されていません。送信されたメールを確認してください。"
						} else {
							msg = "アカウントが有効化されていません。認証メールを再送信しました。確認してください。"
						}
					} else {
						msg = "アカウントが有効化されていません。"
					}

					tmp.Execute(rw, LoginTemp{
						Error:         msg,
						BackURL:       comeback,
						SignupEnabled: settingManager.Get().CanCreateUser,
					})
					return
				}

				sessionID, err := mainDB.SessionAdd(user.Iid)

				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					HttpLog.WithError(err).Error("Session addition failed")

					rw.Write([]byte(ISE500))
				} else {
					SetSession(rw, sessionID)

					RespondRedirection(rw, comeback)
				}
			} else {
				rw.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(rw, BR400)
			}

		})

		return f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/logout"], err = func() (http.Handler, error) {
		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			session := ParseSession(req)

			if session != nil {
				mainDB.SessionRemove(*session)
			}

			cookie := http.Cookie{
				Name:   HttpCookieSession,
				Value:  *session,
				MaxAge: 0,
			}

			http.SetCookie(rw, &cookie)
			RespondRedirection(rw, "/")
		})

		return f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/userinfo"], err = func() (http.Handler, error) {
		tmp, err := template.ParseFiles("./html/userinfo_tmpl.html")

		if err != nil {
			return nil, err
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			user, err := ParseRequestForUserData(req)

			if err != nil {
				if err == ErrUnknownSession {
					RespondRedirection(rw, "/login?comeback=/userinfo")

					return
				}

				HttpLog.WithError(err).Error("ParseRequestForSession failed")

				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(ISE500))

				return
			}

			rw.WriteHeader(http.StatusOK)
			tmp.Execute(rw, user)
		})

		return f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/userinfo/update_password"], err = func() (http.Handler, error) {
		tmp, err := template.ParseFiles("./html/userinfo_update_password_tmpl.html")

		if err != nil {
			return nil, err
		}

		type UpdatePasswordTemplateType struct {
			Success, Error string
			Token          string
		}

		var UPDATEPASSWORDSERVICE = "update_password"
		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			user, err := ParseRequestForUserData(req)

			if err != nil {
				if err == ErrUnknownSession {
					RespondRedirection(rw, "/login?comeback=/userinfo/update_password")

					return
				}

				HttpLog.WithError(err).Error("ParseRequestForUserData failed")

				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(ISE500))

				return
			}

			var val UpdatePasswordTemplateType

			token, err := mainRM.TokenGenerateAndRegisterWithValue(UPDATEPASSWORDSERVICE, time.Duration(settingManager.Get().CSRFTokenExpirationInMinutes)*time.Minute, user.Iid)

			if err != nil {
				DBLog.WithError(err).Error("Token generation and registration failed")

				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(ISE500))

				return
			}
			val.Token = token

			if req.Method == "GET" {
				tmp.Execute(rw, UpdatePasswordTemplateType{
					Token: token,
				})

				return
			} else {
				err := req.ParseForm()

				if err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte(BR400))

					return
				}

				token := req.FormValue("token")
				pass := req.FormValue("password")
				pass2 := req.FormValue("password_conf")

				if len(token) == 0 {
					val.Error = "ワンタイムトークンが無効です。"

					rw.WriteHeader(http.StatusOK)
					tmp.Execute(rw, val)
					return
				}
				ok, iid, err := mainRM.TokenGetAndRemoveInt64(UPDATEPASSWORDSERVICE, token)

				if err != nil {
					DBLog.WithError(err).Error("TokenGetAndRemoveInt64 failed")

					rw.WriteHeader(http.StatusInternalServerError)
					rw.Write([]byte(ISE500))
					return
				}
				if !ok {
					val.Error = "ワンタイムトークンの有効時間が過ぎました。"

					rw.WriteHeader(http.StatusOK)
					tmp.Execute(rw, val)
					return
				}

				if iid != user.Iid {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte(BR400))

					return
				}

				msg := ""
				if len(pass) > 100 || len(pass2) > 100 {
					msg = "パスワードが長すぎます。"
				} else if len(pass) < 5 || len(pass2) < 5 {
					msg = "パスワードが短すぎます。"
				} else if pass != pass2 {
					msg = "パスワードが一致しません。"
				}

				if len(msg) != 0 {
					val.Error = msg

					rw.WriteHeader(http.StatusOK)
					tmp.Execute(rw, val)
					return
				}

				err = mainDB.UserUpdatePassword(iid, pass)

				if err != nil {
					DBLog.WithError(err).Error("UserUpdatePassword failed")
					rw.WriteHeader(http.StatusInternalServerError)
					rw.Write([]byte(ISE500))

					return
				}

				RespondRedirection(rw, "/userinfo")
			}
		})

		return f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/signup/"], err = func() (http.Handler, error) {
		tmp, err := template.ParseFiles("./html/signup_tmpl.html")

		if err != nil {
			return nil, err
		}

		mux := http.NewServeMux()

		var SIGNUPTOKENSERVICE = "signup"
		mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
			u, _ := ParseRequestForUserData(req)

			if u != nil {
				RespondRedirection(rw, "/")

				return
			}

			if !settingManager.Get().CanCreateUser {
				RespondRedirection(rw, "/")

				return
			}

			type TemplateVal struct {
				Success     string
				Error       string
				Token       string
				UserID      string
				UserName    string
				Email       string
				EmailNeeded bool
			}
			var val TemplateVal

			val.EmailNeeded = settingManager.Get().CertificationWithEmail

			token, err := mainRM.TokenGenerateAndRegister(SIGNUPTOKENSERVICE, time.Duration(settingManager.Get().CSRFTokenExpirationInMinutes)*time.Minute)

			if err != nil {
				HttpLog.WithError(err).Error("Token generation or registration failed")

				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(ISE500))

				return
			}
			val.Token = token

			if req.Method == "GET" {
				rw.WriteHeader(http.StatusOK)
				if err := tmp.Execute(rw, val); err != nil {
					HttpLog.WithError(err).Error("text/template execution failed")
				}
			} else if req.Method == "POST" {
				err := req.ParseForm()

				if err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte(BR400))

					return
				}
				wrapFormStr := createWrapFormStr(req)

				token := wrapFormStr("token")
				uid := wrapFormStr("userID")
				userName := wrapFormStr("userName")
				email := wrapFormStr("email")
				pass := wrapFormStr("password")
				pass2 := wrapFormStr("password_conf")

				HttpLog.WithFields(logrus.Fields{
					"token":    token,
					"uid":      uid,
					"userName": userName,
					"email":    email,
					"pass":     pass,
					"pass2":    pass2,
				}).Debug("debug")

				str := func() string {
					arr := make([]string, 0, 10)

					if len(token) == 0 {
						arr = append(arr, "ワンタイムトークンが無効です。")
					} else {
						if ok, err := mainRM.TokenCheckAndRemove(SIGNUPTOKENSERVICE, token); err != nil {
							HttpLog.WithField("token", token).Errorf("Token confirmation failed)")
						} else if !ok {
							arr = append(arr, "ワンタイムトークンの有効時間が過ぎました。")
						}
					}

					if len(uid) > 128 {
						arr = append(arr, "ユーザIDが長すぎます。")
					} else {
						valid := true
						for _, c := range uid {
							if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
								valid = false
							}
						}

						if !valid || len(uid) == 0 {
							arr = append(arr, "ユーザIDが不正です。")
						} else {
							val.UserID = uid
						}
					}

					if len(userName) > 40 {
						arr = append(arr, "ユーザ名が長すぎます。")
					} else if len(userName) == 0 {
						arr = append(arr, "ユーザ名が短すぎます。")
					} else {
						val.UserName = userName
					}

					if len(pass) > 100 || len(pass2) > 100 {
						arr = append(arr, "パスワードが長すぎます。")
					} else if len(pass) < 5 || len(pass2) < 5 {
						arr = append(arr, "パスワードが短すぎます。")
					} else if pass != pass2 {
						arr = append(arr, "パスワードが一致しません。")
					}

					if len(email) > 128 {
						arr = append(arr, "メールアドレスが長すぎます。")
					} else {
						step := 0
						for _, c := range email {
							if step == 0 && c == '@' {
								step++
							} else if step == 1 && c == '.' {
								step++
							}
						}
						if step != 2 {
							arr = append(arr, "メールアドレスが不正です。")
						} else {
							val.Email = email
						}
					}

					return strings.Join(arr, "<br>")
				}()

				if len(str) == 0 {
					gid, err := mainRM.StandardSignupGroupGet()

					if err != nil {
						DBLog.WithError(err).Error("StandardSignupGroupGet failed")
					}
					iid, err := mainDB.UserAdd(uid, userName, pass, NullStringCreate(email), gid, !settingManager.Get().CertificationWithEmail)

					if err != nil {
						if strings.Contains(err.Error(), "Duplicate") {
							arr := make([]string, 0, 2)
							if strings.Contains(err.Error(), "uid") {
								arr = append(arr, "ユーザID")
							}
							if strings.Contains(err.Error(), "user_name") {
								arr = append(arr, "ユーザ名")
							}
							if strings.Contains(err.Error(), "email") {
								arr = append(arr, "メールアドレス")
							}

							val.Error = strings.Join(arr, "と") + "が既に他のアカウントに使用されています。"

							tmp.Execute(rw, val)
							return
						} else {
							DBLog.WithError(err).Error("User addition failed")

							rw.WriteHeader(http.StatusInternalServerError)
							rw.Write([]byte(ISE500))

							return
						}
					}

					if settingManager.Get().CertificationWithEmail {
						err := MailSendConfirmUser(iid, userName, email)

						if err != nil {
							MailLog.WithError(err).Error("Sending mail failed")
							rw.WriteHeader(http.StatusInternalServerError)
							rw.Write([]byte(ISE500))

							return
						}

						var val2 TemplateVal
						val2.Token = val.Token
						val2.Success = "仮登録が完了しました。登録されたメールアドレスに送信されたメールをご確認ください。"

						tmp.Execute(rw, val2)
					} else {
						session, err := mainDB.SessionAdd(iid)

						if err != nil {
							DBLog.WithError(err).WithField("iid", iid).Error("Session addition failed")

							rw.WriteHeader(http.StatusInternalServerError)
							rw.Write([]byte(ISE500))

							return
						}
						SetSession(rw, session)

						RespondRedirection(rw, "/")
					}
				} else {
					val.Error = str

					tmp.Execute(rw, val)
					return
				}
			} else {
				rw.WriteHeader(http.StatusNotImplemented)
				rw.Write([]byte(NI501))
			}
		})

		tmpac, err := template.ParseFiles("./html/signup_account_confirm_tmpl.html")

		if err != nil {
			return nil, err
		}

		type AccountConfirmTemplateType struct {
			OK      bool
			Message string
		}

		mux.HandleFunc("/account_confirm", func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != "GET" {
				rw.WriteHeader(http.StatusMethodNotAllowed)

				return
			}

			err := req.ParseForm()

			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(BR400))

				return
			}

			token := req.FormValue("token")

			if len(token) == 0 {
				rw.WriteHeader(http.StatusForbidden)

				return
			}

			ok, iid, err := mainRM.TokenGetAndRemoveInt64(MAILCONFTOKENSERVICE, token)

			if err != nil {
				DBLog.WithError(err).Error("TokenGetAndRemoveInt64 failed")

				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(ISE500))

				return
			}

			if !ok {
				rw.WriteHeader(http.StatusForbidden)
				if err := tmpac.Execute(rw, AccountConfirmTemplateType{
					OK:      false,
					Message: "URLの有効期限が切れています。メールを再送信するにはログインページにてID/パスワードを送信してください。",
				}); err != nil {
					HttpLog.WithError(err).Error("Execution failed")
				}

				return
			}

			err = mainDB.UserUpdateEnabled(iid, true)

			if err != nil {
				DBLog.WithError(err).Error("UserUpdateEnabled failed")

				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(ISE500))

				return
			}

			rw.WriteHeader(http.StatusOK)
			if err := tmpac.Execute(rw, AccountConfirmTemplateType{
				OK:      true,
				Message: "メールアドレス認証が完了しました。以下のページよりログインしてください。",
			}); err != nil {
				HttpLog.WithError(err).Error("Execution failed")
			}
			return

		})

		return http.StripPrefix("/signup", mux), nil
	}()

	if err != nil {
		return nil, err
	}

	res["/help"], err = func() (*http.HandlerFunc, error) {
		tmp, err := template.ParseFiles("./html/help_tmpl.html")

		if err != nil {
			return nil, err
		}

		type TemplateVal struct {
			Help       html.HTML
			UserName   string
			IsSignedIn bool
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(NF404))

			return

			std, err := ParseRequestForUserData(req)

			var IsSignedIn bool = false
			var Name string
			if err == nil {
				IsSignedIn = true
				Name = std.UserName
			}

			fp, err := os.Open("./html/help.md")

			if err != nil {
				HttpLog.Println(err)
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(ISE500))

				return
			}
			defer fp.Close()

			b, err := ioutil.ReadAll(fp)

			if err != nil {
				HttpLog.Println(err)
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(ISE500))

				return
			}

			unsafe := blackfriday.MarkdownCommon(b)
			page := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

			tmp.Execute(rw, TemplateVal{html.HTML(string(page)), Name, IsSignedIn})
		})

		return &f, nil
	}()

	if err != nil {
		return nil, err
	}

	// Debug request
	res["/admin"], err = func() (*http.HandlerFunc, error) {
		tmp, err := template.ParseFiles("./html/admin_tmpl.html")

		if err != nil {
			return nil, err
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			err := req.ParseForm()

			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(ISE500))

				return
			}

			std, err := ParseRequestForUserData(req)

			if err != nil || std.Gid != 0 {
				RespondRedirection(rw, "/")

				return
			}

			wrapFormStr := createWrapFormStr(req)

			/*type TeplateVal struct {
				ReCAPTCHA bool
				Msg string
				UserID string
				UserName string
				Email string
				Password string
				ReCAPTCHASite string
			}*/

			if req.Method == "GET" {
				tmp.Execute(rw, map[string]string{"UserName": std.UserName})
			} else {
				id := wrapFormStr("id")
				name := wrapFormStr("name")
				pass := wrapFormStr("password")

				if len(id) == 0 || len(name) == 0 || len(pass) == 0 {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte(BR400))

					return
				}

				_, err := mainDB.UserAdd(id, name, pass, NullStringCreate(id+"@hoge.com"), 1, true)

				if err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte("BadRequest: " + err.Error()))
					return
				}

				RespondRedirection(rw, "/admin")
			}
		})

		return &f, nil
	}()

	if err != nil {
		return nil, err
	}

	return res, nil
}
