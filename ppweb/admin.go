package main

import (
	"net/http"
	"text/template"

	"context"

	"github.com/cs3238-tsuzu/popcon-sc/setting"
	"github.com/cs3238-tsuzu/popcon-sc/types"
	muxlib "github.com/gorilla/mux"
)

func AdminHandler() (http.Handler, error) {
	mux := muxlib.NewRouter()
	handler := func(rw http.ResponseWriter, req *http.Request) {
		std, err := ParseRequestForSession(req)

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}

		if !std.IsSignedIn {
			RespondRedirection(rw, "/login?comeback=/admin/")

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

	ProcessUntilError(
		func() error {
			tmpl, err := template.New("").ParseFiles("./html/admin/index_tmpl.html")

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
			tmpl, err := template.New("").ParseFiles("./html/admin/general_tmpl.html")

			if err != nil {
				return err
			}

			type TemplateVal struct {
				UserName string
				Setting  *ppconfiguration.Structure
				Groups   []Group
			}

			mux.HandleFunc("/general", func(rw http.ResponseWriter, req *http.Request) {
				std := sessionTemplateData(req)

				var val TemplateVal
				val.UserName = std.UserName
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
			})

			return nil
		})

	return http.HandlerFunc(handler), nil
}
