package main

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"time"

	"sync"

	"io/ioutil"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/popcon-sc/types"
	"github.com/k0kubun/pp"
	mgo "gopkg.in/mgo.v2"
)

func InitRemoveFile(mux *http.ServeMux, fin <-chan bool, wg *sync.WaitGroup) error {
	wg.Add(1)
	logger := logrus.WithField("feature", "reomve_file")

	var Duration = 5 * time.Minute

	mongo := os.Getenv("PP_MONGO_ADDR")

	session, err := mgo.Dial(mongo)

	if err != nil {
		return err
	}

	db := session.DB("")

	type FileInfo struct {
		Category string
		Path     string
		Time     time.Time
	}

	ch := make(chan FileInfo, 100)
	b, err := ioutil.ReadFile("remove_file_backup.txt")
	var results []FileInfo

	if err == nil {
		err = json.Unmarshal(b, &results)
		if err == nil {
			for i := range results {
				ch <- results[i]
			}
		}
	}

	var status int32
	atomic.StoreInt32(&status, 0)

	results = make([]FileInfo, 100)
	go func() {
		for {
			select {
			case <-fin:
				atomic.AddInt32(&status, 1)
				close(ch)

				idx := 0
				for fi := range ch {
					results[idx] = fi
					idx++
				}

				if b, err := json.Marshal(results); err != nil {
					ioutil.WriteFile("remove_file_backup.txt", b, 0777)
				}

				wg.Done()
				return
			case fi, ok := <-ch:

				if !ok {
					return
				}
				logger.WithField("category", fi.Category).WithField("path", fi.Path).Info("The unnecessary file will be removed.")
				time.Sleep(fi.Time.Add(Duration).Sub(time.Now()))

				err := db.GridFS(fi.Category).Remove(fi.Path)

				if err != nil {
					logger.WithError(err).Error("Remove error")
				} else {
					logger.WithField("category", fi.Category).WithField("path", fi.Path).Info("The unnecessary file was removed.")
				}
			}
		}
	}()

	mux.HandleFunc("/remove_file", func(rw http.ResponseWriter, req *http.Request) {
		if atomic.LoadInt32(&status)%2 == 1 {
			return
		}
		err := req.ParseForm()

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}
		if err := req.ParseForm(); err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}
		fmt.Println(pp.Sprint(req))
		category := req.FormValue("category")
		path := req.FormValue("path")

		ch <- FileInfo{
			Category: category,
			Path:     path,
			Time:     time.Now(),
		}
		logger.WithField("category", category).WithField("path", path).Info("The unnecessary file was pushed into the queue.")

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("OK"))
	})

	return nil
}
