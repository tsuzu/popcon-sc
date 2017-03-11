package main

import (
	"time"

	"io/ioutil"

	"crypto/sha256"

	"gopkg.in/mgo.v2"
)

// Category for MongoDB GridFS
var FS_CATEGORY_SUBMISSION = "submission_code"
var FS_CATEGORY_TESTCASE_SUMMARY = "testcase_summary"
var FS_CATEGORY_TESTCASE_INOUT = "testcase_inout"
var FS_CATEGORY_PROBLEM_STATEMENT = "problem_statement"

var mainFS *MongoFSManager

type MongoFSManager struct {
	session             *mgo.Session
	db                  *mgo.Database
	TestcaseFileBaseTag string
}

func NewMongoFSManager(addr string) (*MongoFSManager, error) {

	cnt := 0
	const RetryingMax = 1000
	var err error

RETRY:
	if cnt != 0 {
		FSLog.Info("Waiting for MongoDB Server Launching...", err.Error())
		time.Sleep(3 * time.Second)
	}
	cnt++

	session, err := mgo.Dial(addr)

	if err != nil {
		if cnt > RetryingMax {
			return nil, err
		}

		goto RETRY
	}

	db := session.DB("")

	b := sha256.Sum256([]byte(time.Now().String()))

	return &MongoFSManager{
		session:             session,
		db:                  db,
		TestcaseFileBaseTag: string(b[:]),
	}, err
}

func (mfs *MongoFSManager) Open(category, path string) (*mgo.GridFile, error) {
	gf, err := mfs.db.GridFS(category).Create(path)

	if err != nil {
		return nil, err
	}

	return gf, nil
}

func (mfs *MongoFSManager) OpenOnly(category, path string) (*mgo.GridFile, error) {
	gf, err := mfs.db.GridFS(category).Open(path)

	if err != nil {
		return nil, err
	}

	return gf, nil
}

func (mfs *MongoFSManager) Read(category, path string) ([]byte, error) {
	gf, err := mfs.OpenOnly(category, path)

	if err != nil {
		return nil, err
	}
	defer gf.Close()

	b, err := ioutil.ReadAll(gf)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func (mfs *MongoFSManager) Write(category, path string, content []byte) error {
	gf, err := mfs.Open(category, path)

	if err == nil {
		return err
	}
	defer gf.Close()

	_, err = gf.Write(content)

	return err
}

func (mfs *MongoFSManager) Ping() error {
	return mfs.session.Ping()
}
