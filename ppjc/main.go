package main

import (
	"os"
	"sync"
	"time"

	"net/http"

	"strconv"

	"encoding/json"

	"io"

	"sync/atomic"

	"context"

	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/chan-utils"
	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
	"github.com/cs3238-tsuzu/popcon-sc/lib/filesystem"
	"github.com/cs3238-tsuzu/popcon-sc/lib/redis"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/cs3238-tsuzu/popcon-sc/ppjc/types"
	mux "github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2"
)

func main() {
	InitProgramExitedNotifier()

	GeneralSetting.RankingRunningTerm = 10
	GeneralSetting.SavingTerm = 5

	token := os.Getenv("PP_TOKEN")
	dbAddr := os.Getenv("PP_MYSQL_ADDR")
	debugMode := os.Getenv("PP_DEBUG_MODE") == "1"
	microServices := os.Getenv("PP_MS_ADDR")
	mongoAddr := os.Getenv("PP_MONGO_ADDR")
	redisAddr := os.Getenv("PP_REDIS_ADDR")
	redisPass := os.Getenv("PP_REDIS_PASS")

	InitLogger(os.Stdout, debugMode)

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

	router.HandleFunc("/contests/{cid}/ranking", func(rw http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}
		vars := mux.Vars(req)

		cid, err := strconv.ParseInt(vars["cid"], 10, 64)
		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)
			return
		}
		limit, err := strconv.ParseInt(req.FormValue("limit"), 10, 64)
		if err != nil {
			limit = -1
		}
		offset, err := strconv.ParseInt(req.FormValue("offset"), 10, 64)
		if err != nil {
			offset = -1
		}

		rows, err := dm.RankingGetAll(cid, offset, limit)

		if err != nil {
			HTTPLog().WithError(err).Error("RankingGetAll() error")
			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

			return
		}

		b, _ := json.Marshal(rows)

		rw.Header().Set("Content-Type", "application/json")
		sctypes.ResponseTemplateWrite(http.StatusOK, rw)
		rw.Write(b)
	})

	router.HandleFunc("/contests/{cid}/problems/add", func(rw http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}
		vars := mux.Vars(req)

		cid, err := strconv.ParseInt(vars["cid"], 10, 64)

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)
			return
		}
		pid, err := strconv.ParseInt(req.FormValue("pid"), 10, 64)
		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)
			return
		}

		if err := dm.RankingProblemAdd(cid, pid); err != nil {
			HTTPLog().WithError(err).Error("RankingProblemAdd() error")
			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)
			return
		}
	})
	router.HandleFunc("/contests/{cid}/problems/delete", func(rw http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}
		vars := mux.Vars(req)

		cid, err := strconv.ParseInt(vars["cid"], 10, 64)

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)
			return
		}
		pid, err := strconv.ParseInt(req.FormValue("pid"), 10, 64)
		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)
			return
		}

		if err := dm.RankingProblemDelete(cid, pid); err != nil {
			HTTPLog().WithError(err).Error("RankingProblemDelete() error")
			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)
			return
		}
	})

	router.HandleFunc("/contests/{cid}/new", func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)

		cid, err := strconv.ParseInt(vars["cid"], 10, 64)

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)
			return
		}

		if err := dm.RankingAutoMigrate(cid); err != nil {
			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)

			return
		}
	})

	router.HandleFunc("/contests/{cid}/submit", func(rw http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}

		cid, err := strconv.ParseInt(mux.Vars(req)["cid"], 10, 64)
		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)
			return
		}
		sid, err := strconv.ParseInt(req.FormValue("sid"), 10, 64)
		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)
			return
		}

		err = rm.JudgeQueuePush(cid, sid)

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)
			return
		}

		sctypes.ResponseTemplateWrite(http.StatusOK, rw)
	})

	router.HandleFunc("/file_download", func(rw http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()

		if err != nil {
			sctypes.ResponseTemplateWrite(http.StatusBadRequest, rw)

			return
		}

		category := req.FormValue("category")
		name := req.FormValue("path")

		if len(category) == 0 || len(name) == 0 {
			sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)
			return
		}

		fp, err := fsm.OpenOnly(category, name)

		if err != nil {
			if err == mgo.ErrNotFound {
				sctypes.ResponseTemplateWrite(http.StatusNotFound, rw)
				return
			}

			FSLog().WithError(err).Error("OpenOnly() error")

			sctypes.ResponseTemplateWrite(http.StatusInternalServerError, rw)
			return
		}
		defer fp.Close()

		rw.Header().Set("Content-Length", strconv.FormatInt(fp.Size(), 10))
		rw.Header().Set("Content-Type", "text/plain")

		sctypes.ResponseTemplateWrite(http.StatusOK, rw)
		io.Copy(rw, fp)
	})

	upgrader := websocket.Upgrader{}
	router.HandleFunc("/workers/ws/polling", func(rw http.ResponseWriter, req *http.Request) {
		parallelJudge, err := strconv.ParseInt(req.Header.Get("Popcon-Parallel-Judge"), 10, 64)

		if err != nil || parallelJudge <= 0 {
			parallelJudge = 1
		}
		conn, err := upgrader.Upgrade(rw, req, nil)

		if err != nil {
			HTTPLog().WithError(err).Error("Upgrade() for websocket error")

			return
		}

		defer conn.Close()

		var availableThread int64
		atomic.StoreInt64(&availableThread, parallelJudge)
		trigger := chanUtils.NewTrigger()
		closed := chanUtils.NewExitedNotifier()
		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			for {
				if val := atomic.LoadInt64(&availableThread); val > 0 {
					atomic.AddInt64(&availableThread, -1)
					ctx, canceler := context.WithCancel(context.Background())
					fin := programExitedNotifier.TriggerOrCancel(func() {
						canceler()
					})

					cid, sid, err := rm.JudgeQueuePopBlockingWithContext(5, ctx)
					fin()

					if err != nil {
						if err == context.Canceled || err == context.DeadlineExceeded {
							return
						}

						logrus.WithError(err).Error("JudgeQueuePopBlockingWithContext() error")

						atomic.AddInt64(&availableThread, 1)
						time.Sleep(5 * time.Second)

						continue
					}

					var info ppjctypes.JudgeInformation

					if err := dm.BeginDM(func(dm *database.DatabaseManager) error {
						sm, err := dm.SubmissionFind(cid, sid)

						if err != nil {
							return err
						}
						info.Submission = *sm

						prob, err := dm.ContestProblemFind(cid, sm.Pid)

						if err != nil {
							return err
						}
						info.Problem = *prob

						cases, scores, err := prob.LoadTestCases()

						if err != nil {
							return err
						}
						info.Cases, info.Scores = cases, scores

						return nil
					}); err != nil {
						logrus.WithField("cid", cid).WithField("sid", sid).WithError(err).Error("Get information for judge error")

						atomic.AddInt64(&availableThread, 1)
						time.Sleep(5 * time.Second)

						continue
					}

					jid, err := rm.JudgeIDGet()

					if err != nil {
						logrus.WithError(err).Error("JudgeIDGet() error")

						atomic.AddInt64(&availableThread, 1)
						time.Sleep(5 * time.Second)

						continue
					}

					info.Submission.Jid = jid

					if err := conn.WriteJSON(info); err != nil {
						logrus.WithField("cid", cid).WithField("sid", sid).WithError(err).Error("WriteJSON error")

						atomic.AddInt64(&availableThread, 1)
						time.Sleep(5 * time.Second)

						continue
					}
				}

				if atomic.LoadInt64(&availableThread) > 0 {
					continue
				}
				select {
				case <-trigger:
				case <-programExitedNotifier.Channel:
					return
				case <-closed.Channel:
					return
				}
			}
		}()

		var msg ppjctypes.PollingMessage
		for {
			err := conn.ReadJSON(&msg)

			if err != nil {
				closed.Finish()
				break
			}

			switch msg {
			case ppjctypes.JudgeOneFinished:
				atomic.AddInt64(&availableThread, 1)
			}
		}

		wg.Wait()
	})

	//	router.HandleFunc("/contests/{cid}/", f func(http.ResponseWriter, *http.Request))

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get("Authentication")

		if auth != token {
			sctypes.ResponseTemplateWrite(http.StatusForbidden, rw)

			return
		}

		router.ServeHTTP(rw, req)
	})

	if err := http.ListenAndServe(":80", handler); err != nil {
		HTTPLog().WithError(err).Fatal("ListenAndServe() error")

		return
	}

	return
}
