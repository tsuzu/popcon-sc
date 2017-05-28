package main

import (
	"flag"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
	"github.com/cs3238-tsuzu/popcon-sc/lib/filesystem"
	"github.com/cs3238-tsuzu/popcon-sc/lib/redis"
	"github.com/cs3238-tsuzu/popcon-sc/ppjc/client"
	"github.com/facebookgo/grace/gracehttp"
	gorilla "github.com/gorilla/handlers"
	"github.com/sebest/xff"
)

var mainDB *database.DatabaseManager
var mainRM *redis.RedisManager
var mainFS *fs.MongoFSManager
var ppjcClient *ppjc.Client

func main() {
	var err error
	// 標準時
	time.Local = Location

	pprof := os.Getenv("PP_PPROF") == "1"

	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		flag.PrintDefaults()

		return
	}

	if pprof {
		l, err := net.Listen("tcp", ":54345")

		if err != nil {
			logrus.Fatal(err.Error())

			return
		}
		logrus.Info("pprof server is listening on %s\n", l.Addr())
		go http.Serve(l, nil)
	}
	environmentalSetting.dbAddr = os.Getenv("PP_MYSQL_ADDR")
	environmentalSetting.mongoAddr = os.Getenv("PP_MONGO_ADDR")
	environmentalSetting.redisAddr = os.Getenv("PP_REDIS_ADDR")
	environmentalSetting.redisPass = os.Getenv("PP_REDIS_PASS")
	environmentalSetting.judgeControllerAddr = os.Getenv("PP_JC_ADDR")
	environmentalSetting.microServicesAddr = os.Getenv("PP_MS_ADDR")
	environmentalSetting.internalToken = os.Getenv("PP_TOKEN")
	environmentalSetting.debugMode = os.Getenv("PP_DEBUG_MODE") == "1"

	ppjcClient, err = ppjc.NewClient(environmentalSetting.judgeControllerAddr, environmentalSetting.internalToken)

	if err != nil {
		panic(err)
	}

	// ロガー作成
	InitLogger(os.Stdout, environmentalSetting.debugMode)

	// Redis
	mainRM, err = redis.NewRedisManager(environmentalSetting.redisAddr, environmentalSetting.redisPass, DBLog)
	if err != nil {
		DBLog().WithError(err).Fatal("Redis initialization failed")
	}
	defer mainRM.Close()

	// MongoDB
	mainFS, err = fs.NewMongoFSManager(environmentalSetting.mongoAddr, environmentalSetting.microServicesAddr, environmentalSetting.internalToken, mainRM, FSLog)

	if err != nil {
		FSLog().WithError(err).Fatal("MongoDB FS initialization failed")
	}
	defer mainFS.Close()

	// MySQL Database
	mainDB, err = database.NewDatabaseManager(environmentalSetting.dbAddr, environmentalSetting.debugMode, mainFS, mainRM, DBLog)

	if err != nil {
		DBLog().WithError(err).Fatal("Database initialization failed")
	}
	database.SetDefaultManager(mainDB)
	defer mainDB.Close()

	userCnt, err := mainDB.UserCount()

	if err != nil {
		DBLog().Println("Failed to count the users", err.Error())

		return
	}

	if userCnt == 0 {
		if !CreateAdminUserAutomatically() {
			if cnt, err := mainDB.UserCount(); cnt == 0 || err != nil {
				DBLog().Println("Admin user creation failed.")

				return
			}
		}
	}

	mux := http.NewServeMux()
	handlers, err := CreateHandlers()

	if err != nil {
		HttpLog().WithError(err).Fatal("CreateHandlers() error")
	}

	for k, v := range handlers {
		mux.Handle(k, v)
	}

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	mux.HandleFunc("/favicon.ico", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Location", "/static/popcon.ico")
		rw.WriteHeader(http.StatusMovedPermanently)
	})

	xffh, err := xff.Default()

	if err != nil {
		HttpLog().Fatal(err)
	}

	logger := gorilla.LoggingHandler(
		NewCustomizedWriter(
			func(b []byte) (int, error) {
				HttpLog().Info(string(b))

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
		Addr:           ":80",
		MaxHeaderBytes: 1 << 20,
		Handler:        xssProtector,
	}

	// SSL should be provided at the load balancer
	/*	if len(environmentalSetting.CertFilePath) != 0 && len(environmentalSetting.KeyFilePath) != 0 {
		cer, err := tls.LoadX509KeyPair(environmentalSetting.CertFilePath, environmentalSetting.KeyFilePath)
		if err != nil {
			HttpLog().Fatal(err)
		}

		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		server.TLSConfig = config
	}*/

	err = gracehttp.Serve(server)

	if err != nil {
		HttpLog().Fatal(err)
	}
}
