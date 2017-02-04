package main

import (
	"encoding/json"
	"flag"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"io/ioutil"

	"path/filepath"

	"crypto/tls"

	"github.com/Sirupsen/logrus"
	"github.com/facebookgo/grace/gracehttp"
	gorilla "github.com/gorilla/handlers"
	"github.com/sebest/xff"
)

func main() {
	time.Local = Location

	settingFile := os.Getenv("PP_SETTING")
	pprof := os.Getenv("PP_PPROF")

	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		flag.PrintDefaults()

		return
	}

	if len(pprof) != 0 {
		l, err := net.Listen("tcp", pprof)

		if err != nil {
			logrus.Fatal(err.Error())

			return
		}
		logrus.Info("pprof server is listening on %s\n", l.Addr())
		go http.Serve(l, nil)
	}

	fp, err := os.OpenFile(settingFile, os.O_RDONLY, 0664)

	if err != nil {
		b, _ := json.Marshal(Setting{})

		logrus.Info("Json Sample: ", string(b))
		ioutil.WriteFile(settingFile, b, 0660)

		os.Exit(1)

		return
	}

	dec := json.NewDecoder(fp)

	var setting Setting
	err = dec.Decode(&setting)

	if err != nil {
		logrus.Fatal("Syntax error: ", err)

		return
	}

	setting.dbAddr = os.Getenv("PP_MYSQL_ADDR")
	setting.redisAddr = os.Getenv("PP_REDIS_ADDR")
	setting.redisPass = os.Getenv("PP_REDIS_PASS")
	setting.judgeControllerAddr = os.Getenv("PP_JC_ADDR")
	setting.internalToken = os.Getenv("PP_TOKEN")
	setting.listeningEndpoint = os.Getenv("PP_LISTEN")
	setting.dataDirectory = os.Getenv("PP_DATA_DIR")

	if setting.CertificationWithEmail && (setting.SendmailCommand == nil || len(setting.SendmailCommand) == 0) {
		logrus.Fatal("SendmailCommand mustn't be empty in the setting")
	}

	if len(setting.listeningEndpoint) == 0 {
		setting.listeningEndpoint = ":80"
	}

	settingManager.Set(setting)

	lo, err := NewLogMultipleOutput(settingManager.Get().LogFile)

	if err != nil {
		logrus.Fatal(err.Error())
	}
	// ロガー作成
	CreateLogger(lo)

	dir := settingManager.Get().dataDirectory

	SubmissionDir = filepath.Join(dir, SubmissionDir)
	if err := os.MkdirAll(SubmissionDir, 0770); err != nil {
		HttpLog.Fatalf("Creation of SubmissionDir(%s) failed(error: %s)", SubmissionDir, err.Error())
	}
	ContestDir = filepath.Join(dir, ContestDir)
	if err := os.MkdirAll(ContestDir, 0770); err != nil {
		HttpLog.Fatalf("Creation of ContestDir(%s) failed(error: %s)", ContestDir, err.Error())
	}
	ContestProblemDir = filepath.Join(dir, ContestProblemDir)
	if err := os.MkdirAll(ContestProblemDir, 0770); err != nil {
		HttpLog.Fatalf("Creation of ContestProblemDir(%s) failed(error: %s)", ContestProblemDir, err.Error())
	}

	mainRM, err = NewRedisManager(setting.redisAddr, setting.redisPass)

	if err != nil {
		DBLog.Fatal(err.Error())
	}

	mainDB, err = NewDatabaseManager(os.Getenv("PP_WAIT_DB") != "")

	if err != nil {
		DBLog.Fatal(err.Error())
	}

	userCnt, err := mainDB.UserCount()

	if err != nil {
		DBLog.Println("Failed to count the users", err.Error())

		return
	}

	if userCnt == 0 {
		if !CreateAdminUserAutomatically() {
			if cnt, err := mainDB.UserCount(); cnt == 0 || err != nil {
				DBLog.Println("Admin user creation failed.")

				return
			}
		}
	}

	mux := http.NewServeMux()
	handlers, err := CreateHandlers()

	if err != nil {
		HttpLog.Fatal(err)
	}

	for k, v := range handlers {
		mux.Handle(k, v)
	}

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	//mux.Handle("/judge", JudgeTransfer{})
	mux.HandleFunc("/favicon.ico", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Location", "/static/popcon.ico")
		rw.WriteHeader(http.StatusMovedPermanently)
	})

	xffh, err := xff.Default()

	if err != nil {
		HttpLog.Fatal(err)
	}

	logger := gorilla.LoggingHandler(
		NewCustomizedWriter(
			func(b []byte) (int, error) {
				HttpLog.Info(string(b))

				return len(b), nil
			},
		),
		xffh.Handler(mux),
	)

	xssProtector := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("X-XSS-Protection", "1")
		logger.ServeHTTP(rw, req)
	})

	// Should use TLS
	server := &http.Server{
		Addr:           settingManager.Get().listeningEndpoint,
		MaxHeaderBytes: 1 << 20,
		Handler:        xssProtector,
	}

	if len(setting.CertFilePath) != 0 && len(setting.KeyFilePath) != 0 {
		cer, err := tls.LoadX509KeyPair(setting.CertFilePath, setting.KeyFilePath)
		if err != nil {
			HttpLog.Fatal(err)
		}

		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		server.TLSConfig = config
	}

	err = gracehttp.Serve(server)

	if err != nil {
		HttpLog.Fatal(err)
	}
}
