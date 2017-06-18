package main

import (
	"context"
	"flag"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
	"github.com/cs3238-tsuzu/popcon-sc/lib/filesystem"
	"github.com/cs3238-tsuzu/popcon-sc/lib/redis"
	"github.com/cs3238-tsuzu/popcon-sc/lib/traefik_consul"
	"github.com/cs3238-tsuzu/popcon-sc/ppjc/client"
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
	traefikRegistrationFlag := flag.Bool("enable-traefik-registration", false, "Register/Unregister the address of this server to consul for traefik automatically.")
	debugFlag := flag.Bool("debug", false, "Debug logging")

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
	environmentalSetting.debugMode = *debugFlag

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

	var traefikShutdown func()
	func() {
		if !*traefikRegistrationFlag {
			return
		}
		addr := os.Getenv("PP_CONSUL_ADDR")
		prefix := os.Getenv("PP_TRAEFIK_PREFIX")
		if len(prefix) == 0 {
			prefix = "traefik"
		}

		client, err := traefikConsul.NewClient(prefix, addr)
		if err != nil {
			HttpLog().WithError(err).Error("traefikConsul.NewClient() error")

			return
		}

		var advertiseAddr string
		if addr := os.Getenv("PP_ADVERTISE_ADDR"); len(addr) != 0 {
			u, err := url.Parse(addr)
			if err != nil {
				HttpLog().WithError(err).Errorf("URL format is illegal. The default one will be used.", addr)

			} else {
				if u.Scheme == "" {
					u.Scheme = "http"
				}
				advertiseAddr = u.String()
			}
		} else if iface := os.Getenv("PP_IFACE"); len(iface) != 0 {
			addr, err := traefikConsul.IPAddressFromIface(iface)

			if err != nil {
				HttpLog().WithError(err).Errorf("Network interface(%s) was not found. The default one will be used.", iface)
			} else {
				advertiseAddr = "http://" + addr
			}

		}

		if len(advertiseAddr) == 0 {
			addr, err := traefikConsul.DefaultIPAddress()

			if err != nil {
				HttpLog().WithError(err).Error("traefikConsul.DefaultIPAddress() error")
				return
			}

			advertiseAddr = "http://" + addr
		}

		backend := os.Getenv("PP_TRAEFIK_BACKEND")
		if len(backend) == 0 {
			backend = "backend1"
		}

		host, _ := os.Hostname()
		const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
		rgen := rand.New(rand.NewSource(time.Now().UnixNano()))
		result := make([]byte, 6)
		for i := range result {
			result[i] = chars[rgen.Intn(len(chars))]
		}
		serverName := "ppweb-" + host + "-" + string(result)

		if b, err := ioutil.ReadFile("./traefik_config_backup"); err == nil {
			if err := client.RestoreBackup(backend, serverName, b); err != nil {
				HttpLog().WithError(err).Error("RestoreBackup() error")
			}
		}
		var wg sync.WaitGroup
		once := sync.Once{}
		shutdown := func() {
			wg.Wait()
			backup, err := client.BackupBackend(backend, serverName)

			if err != nil {
				HttpLog().WithError(err).Error("BackupBackend() error")

				return
			}

			ioutil.WriteFile("./traefik_config_backup", backup, 0777)

			if err := client.DeleteBackend(backend, serverName); err != nil {
				HttpLog().WithError(err).Error("BackupBackend() error")

				return
			}
		}
		traefikShutdown = func() {
			once.Do(shutdown)
		}

		wg.Add(1)
		go func() {
			if err := client.RegisterNewBackend(backend, serverName, advertiseAddr); err != nil {
				HttpLog().WithError(err).Error("RegisterNewBackend() error")
			}

			wg.Done()
		}()
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-signalChan

		traefikShutdown()
		time.Sleep(1 * time.Second)
		ctx, f := context.WithTimeout(context.Background(), 60*time.Second)
		server.Shutdown(ctx)
		f()
	}()
	if err := server.ListenAndServe(); err != nil {
		if err != nil {
			HttpLog().Error(err)
		}
	}
	traefikShutdown()
}
