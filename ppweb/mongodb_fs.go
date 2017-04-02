package main

import (
	"io"
	"time"

	ppms "github.com/cs3238-tsuzu/popcon-sc/ppms/client"

	"io/ioutil"

	"crypto/sha256"

	"strconv"

	"strings"

	"gopkg.in/mgo.v2"
)

// Category for MongoDB GridFS
var FS_CATEGORY_SUBMISSION = "submission_code"
var FS_CATEGORY_TESTCASE_SUMMARY = "testcase_summary"
var FS_CATEGORY_TESTCASE_INOUT = "testcase_inout"
var FS_CATEGORY_CONTEST_DESCRIPTION = "contest_description"
var FS_CATEGORY_PROBLEM_STATEMENT = "problem_statement"
var FS_CATEGORY_PROBLEM_CHECKER = "problem_checker"

var mainFS *MongoFSManager

type MongoFSManager struct {
	session             *mgo.Session
	db                  *mgo.Database
	msClient            *ppms.Client
	TestcaseFileBaseTag string
}

func NewMongoFSManager(addr, msaddr, token string) (*MongoFSManager, error) {

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

	client, err := ppms.NewClient(msaddr, token)

	if err != nil {
		return nil, err
	}

	return &MongoFSManager{
		session:             session,
		db:                  db,
		msClient:            client,
		TestcaseFileBaseTag: string(b[:]),
	}, err
}

func (mfs *MongoFSManager) Close() {
	mfs.session.Close()
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

func (mfs *MongoFSManager) Remove(category, path string) error {
	return mfs.db.GridFS(category).Remove(path)
}

func (mfs *MongoFSManager) RemoveID(category string, id interface{}) error {
	return mfs.db.GridFS(category).RemoveId(id)
}

func (mfs *MongoFSManager) Ping() error {
	return mfs.session.Ping()
}

func (mfs *MongoFSManager) RemoveLater(category, path string) error {
	if len(path) == 0 {
		return nil
	}

	return mfs.msClient.RemoveFile(category, path)
}

func (mfs *MongoFSManager) CreateFilePath(category string, version int64) string {
	return category + "_" + strconv.FormatInt(version, 10) + ".txt"
}

func (mfs *MongoFSManager) FileUpdate(category, oldName, newData string) (string, error) {
	FSLog.Warn("FileUpdate() is deprecated. Should use FileSecureUpdate()")

	f, n, e := mfs.FileSecureUpdate(category, oldName, newData)

	if f != nil {
		f()
	}

	return n, e
}

func (mfs *MongoFSManager) FileSecureUpdateWithReader(category, oldName string, reader io.Reader) (func(), string, error) {
	removeSecurely := func() {}

	if len(oldName) != 0 {
		removeSecurely = func() {
			err := mfs.msClient.RemoveFile(category, oldName)

			if err != nil {
				FSLog.WithError(err).Error("RemoveFile error")
			}
		}
	}

	fid, err := mainRM.UniqueFileID(category)

	if err != nil {
		return nil, "", err
	}
	newName := mfs.CreateFilePath(category, fid)

	fp, err := mainFS.Open(category, newName)

	if err != nil {
		return nil, "", err
	}
	defer fp.Close()

	_, err = io.Copy(fp, reader)

	if err != nil {
		return nil, "", err
	}

	return removeSecurely, newName, nil
}

func (mfs *MongoFSManager) FileSecureUpdate(category, oldName, newData string) (func(), string, error) {
	return mfs.FileSecureUpdateWithReader(category, oldName, strings.NewReader(newData))
}
