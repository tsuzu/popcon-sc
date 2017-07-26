package main

import (
	"net/http"
	"os"
	"sync/atomic"
	"sync"
	"encoding/json"
	"time"
	"bytes"
	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	mgo "gopkg.in/mgo.v2"
	"github.com/boltdb/bolt"
)
const Bucket = "Files"

func InitRemoveFile(mux *http.ServeMux, fin <-chan bool, wg *sync.WaitGroup) error {
	wg.Add(1)
	logger := func() *logrus.Entry { return logrus.WithField("feature", "reomve_file") }

	var Duration = 5 * time.Minute

	mongo := os.Getenv("PP_MONGO_ADDR")
	dbFile := os.Getenv("PP_BOLTDB_DB")

	if dbFile == "" {
		dbFile = "/root/remove_file.db"
	}

	boltdb, err := bolt.Open("my.db", 0666, nil)
	if err != nil {
		logger().WithError(err).Error(err)

		return err
	}

	boltdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(Bucket))

		return err
	})

	session, err := mgo.Dial(mongo)

	if err != nil {
		return err
	}

	db := session.DB("")

	type FileInfo struct {
		Category string
		Path     string
	}

	var status int32
	atomic.StoreInt32(&status, 0)

	go func() {
		defer boltdb.Close()
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-fin:
				atomic.AddInt32(&status, 1)
				wg.Done()
			case <-ticker.C:
				files := make([]FileInfo, 0, 10)
				boltdb.Update(func(tx *bolt.Tx) error {
					c := tx.Bucket([]byte(Bucket)).Cursor()

					min := []byte(time.Unix(0, 0).Format(time.RFC3339))
					max := []byte(time.Now().Format(time.RFC3339))

					for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
						c.Delete()

						var f FileInfo
						json.Unmarshal(v, &f)

						files = append(files, f)
					}
					return nil
				})

				for i := range files {
					err := db.GridFS(files[i].Category).Remove(files[i].Path)

					if err != nil {
						logger().WithError(err).Error("Remove error")
					} else {
						logger().WithField("category", files[i].Category).WithField("path", files[i].Path).Info("The unnecessary file was removed.")
					}
				}
			}
		}
	}()

	mux.HandleFunc("/remove_file", func(rw http.ResponseWriter, req *http.Request) {
		if atomic.LoadInt32(&status)%2 == 1 {
			rw.WriteHeader(http.StatusServiceUnavailable)
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

		category := req.FormValue("category")
		path := req.FormValue("path")

		json, _ := json.Marshal(FileInfo{
			Category: category,
			Path: path,
		})
		if err := boltdb.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(Bucket))
			err := b.Put([]byte(time.Now().Add(Duration).Format(time.RFC3339) + "_" + category + "_" + path), json)

			return err
		}); err != nil {
			logger().WithError(err).WithField("category", category).WithField("path", path).Error("Failed to push new data")
		}else {
			logger().WithField("category", category).WithField("path", path).Info("The file was pushed into the queue.")
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("OK"))
	})

	return nil
}
