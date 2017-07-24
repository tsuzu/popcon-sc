package main

import (
	"context"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
	"github.com/cs3238-tsuzu/popcon-sc/lib/filesystem"
	"github.com/cs3238-tsuzu/popcon-sc/lib/redis"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	gorilla "github.com/gorilla/handlers"
	mux "github.com/gorilla/mux"
	"github.com/sebest/xff"
)

func main() {
	InitProgramExitedNotifier()

	token := os.Getenv("PP_TOKEN")
	dbAddr := os.Getenv("PP_MYSQL_ADDR")
	debugMode := os.Getenv("PP_DEBUG_MODE") == "1"
	microServices := os.Getenv("PP_MS_ADDR")
	mongoAddr := os.Getenv("PP_MONGO_ADDR")
	redisAddr := os.Getenv("PP_REDIS_ADDR")
	redisPass := os.Getenv("PP_REDIS_PASS")

	InitLogger(os.Stdout, debugMode)

	if os.Getenv("PP_PPROF") == "1" {
		l, err := net.Listen("tcp", ":54345")

		if err != nil {
			logrus.Fatal(err.Error())

			return
		}
		HTTPLog().Info("pprof server is listening on %s\n", l.Addr())
		go http.Serve(l, nil)
	}

	rm, err := redis.NewRedisManager(redisAddr, redisPass, DBLog)

	if err != nil {
		DBLog().WithError(err).Fatal("Redis initialization error")
	}
	defer rm.Close()

	fsm, err := fs.NewMongoFSManager(mongoAddr, microServices, token, rm, FSLog)

	if err != nil {
		DBLog().WithError(err).Fatal("MongoDB FS initialization error")
	}
	defer fsm.Close()

	dm, err := database.NewDatabaseManager(dbAddr, debugMode, fsm, rm, DBLog)

	if err != nil {
		DBLog().WithError(err).Fatal("Database initialization error")
	}

	database.SetDefaultManager(dm)
	defer dm.Close()

	router := mux.NewRouter()

	v1 := &HandlerV1{
		DM:  dm,
		RM:  rm,
		FSM: fsm,
	}

	v1.Route(router)

	//	router.HandleFunc("/contests/{cid}/", f func(http.ResponseWriter, *http.Request))

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get(sctypes.InternalHTTPToken)

		if auth != token {
			sctypes.ResponseTemplateWrite(http.StatusForbidden, rw)

			return
		}

		router.ServeHTTP(rw, req)
	})

	xffh, err := xff.Default()

	if err != nil {
		HTTPLog().WithError(err).Fatal("xff setup error")
	}

	logger := gorilla.LoggingHandler(
		NewCustomizedWriter(
			func(b []byte) (int, error) {
				HTTPLog().Info(string(b))

				return len(b), nil
			},
		),
		xffh.Handler(handler),
	)

	server := &http.Server{
		Addr:           ":80",
		MaxHeaderBytes: 1 << 20,
		Handler:        logger,
	}

	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	go func() {
		<-signalChan

		ctx, f := context.WithTimeout(context.Background(), 60*time.Second)
		server.Shutdown(ctx)
		f()
	}()
	if err := server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			HTTPLog().WithError(err).Fatal("ListenAndServe() error")
		}
		return
	}

	return
}
