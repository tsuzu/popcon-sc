package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cs3238-tsuzu/popcon-sc/ppjc/client"
	"github.com/cs3238-tsuzu/popcon-sc/ppjc/types"
	"github.com/docker/docker/client"
	// TODO: This will be updated to moby/moby/client

	"os/signal"
	"sync"
	"syscall"

	"strings"

	"strconv"

	"io"

	"io/ioutil"

	"encoding/json"

	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
	"github.com/cs3238-tsuzu/popcon-sc/lib/filesystem"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/cs3238-tsuzu/popcon-sc/lib/utility"
)

// SettingsTemplate is a template of a setting json
const SettingsTemplate = `{
    "name": "test-server",
    "parallelism": 2,
    "cpu_usage": 100,
	"auth": "****",
	"languages": {},
}`

type Language struct {
	SourceFileName string   `json:"source_file_name"`
	Compile        bool     `json:"compile_necessity"`
	CompileCmd     []string `json:"compile_command"`
	CompileImage   string   `json:"compile_image"`
	ExecCmd        []string `json:"exec_command"`
	ExecImage      string   `json:"exec_image"`
}

// SettingsInterface is a interface of setting file
// Generated at https://mholt.github.io/json-to-go/
type SettingsInterface struct {
	Name        string              `json:"name"`
	Parallelism int                 `json:"parallelism"`
	CPUUsage    int                 `json:"cpu_usage"`
	Auth        string              `json:"auth"`
	Languages   map[string]Language // string(lid int64)
}

func CreateStringPointer(str string) *string {
	return &str
}

func main() {
	help := flag.Bool("help", false, "Display all options")
	wdir := flag.String("exec-dir", "/tmp/pj", "Directory to execute programs and save files")
	server := flag.String("server", "http://192.168.2.1:8080/", "popcon-sc server address for ppjc node")
	auth := flag.String("auth", "", "authentication token for popcon-sc ppjc")
	parallel := flag.Int64("parallel", 1, "the number of executions in parallel")
	cgroup := flag.String("cgroup", "/sys/fs/cgroup", "cgroup dir")
	docker := flag.String("docker", "unix:///var/run/docker.sock", "docker host path")
	cpuUsage := flag.Int("cpu-util", 100, "restriction of CPU utilization(expressed with per-cent, so must be 1-100)")
	winmacMode := flag.Bool("winmac", false, "whether this is running on Docker for Mac or Windows")
	langSetting := flag.String("lang", "./languages.json", "the path to configuration file of languages(json/yaml)")
	echoLangSettingTemplate := flag.String("echo-lang-setting", "none", "Display the template of configuration of languages(none/json/yaml/etc...)")
	debug := flag.Bool("debug", false, "debug mode")

	flag.Parse()

	if help != nil && *help {
		flag.PrintDefaults()

		return
	}

	if EchoLanguageConfigurationTemplate(os.Stdout, *echoLangSettingTemplate) {
		return
	}

	if *cpuUsage < 1 || *cpuUsage > 100 {
		log.Fatal("--cpu-util option must be [1, 100]")

		return
	}

	cgroupDir = *cgroup

	err := os.MkdirAll(*wdir, 0664)

	if err != nil {
		log.Println(err.Error())

		os.Exit(1)

		return
	}

	if *debug {
		InitLogger(os.Stderr, true)
	}

	headers := map[string]string{"User-Agent": "popcon-judge/v1.00"}

	cli, err = client.NewClient(*docker, "v1.29", nil, headers)

	if err != nil {
		panic(err)
	}

	var lmutex sync.RWMutex
	var languages LanguageConfiguration

	if l, err := LoadLanguageConfiguration(*langSetting); err != nil {
		FSLog().WithError(err).Fatal(*langSetting, "has syntax errors")

		return
	} else {
		languages = l
	}

	// Reload language configurations
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGUSR1)

		for {
			select {
			case <-ch:
				l, err := LoadLanguageConfiguration(*langSetting)
				if err != nil {
					FSLog().WithError(err).Error(*langSetting, "has syntax errors")

					continue
				}

				lmutex.Lock()
				languages = l
				lmutex.Unlock()

				FSLog().WithField("config", l).Info("Succeed in reloading the configuration file of languages")
				// TODO: Add case for ExitedNotifier
			}
		}
	}()

	pclient, err := ppjc.NewClient(*server, *auth)

	if err != nil {
		panic(err)
	}

	ppd, err := NewPPDownloader(pclient, *wdir)

	ppd.RunAutomaticalyDeleter(context.Background(), 30*time.Minute /*TODO: can change by prgoram options*/)

	if err != nil {
		panic(err)
	}

	jinfoChan := make(chan ppjctypes.JudgeInformation, 100)
	ctx, canceller := context.WithCancel(context.Background())

	defer canceller()

	exited, oneFinish, err := pclient.StartWorkersWSPolling(*parallel, jinfoChan, ctx)

	if err != nil {
		panic(err)
	}

	for {
		req, ok := <-jinfoChan

		if !ok {
			return
		}

		go func() {
			j := Judge{}

			code, err := pclient.FileDownload(fs.FS_CATEGORY_SUBMISSION, req.Submission.CodeFile)

			if err != nil {
				pclient.JudgeSubmissionsUpdateResult(
					req.Submission.Cid,
					req.Submission.Sid,
					req.Submission.Jid,
					sctypes.SubmissionStatusInternalError,
					0, 0, 0,
					strings.NewReader("Failed downloading the source code"),
				)

				return
			}
			defer code.Close()

			var checker io.ReadCloser
			if req.Problem.Type == sctypes.JudgeInteractive || req.Problem.Type == sctypes.JudgeRunningCode {
				checker, err = pclient.FileDownload(fs.FS_CATEGORY_PROBLEM_CHECKER, req.Problem.CheckerFile)

				if err != nil {
					pclient.JudgeSubmissionsUpdateResult(
						req.Submission.Cid,
						req.Submission.Sid,
						req.Submission.Jid,
						sctypes.SubmissionStatusInternalError,
						0, 0, 0,
						strings.NewReader("Failed downloading the checker"),
					)

					return
				}
			}
			defer checker.Close()

			lmutex.RLock()
			lang, has := languages[req.Submission.Lang]
			lmutex.RUnlock()

			if !has {
				pclient.JudgeSubmissionsUpdateResult(
					req.Submission.Cid,
					req.Submission.Sid,
					req.Submission.Jid,
					sctypes.SubmissionStatusInternalError,
					0, 0, 0,
					strings.NewReader("Unknown language(with lid:"+strconv.FormatInt(req.Submission.Lang, 10)+")"),
				)

				return
			}

			j.Code = code
			j.Time = req.Problem.Time * 1000      // sec -> ms
			j.Mem = req.Problem.Mem * 1000 * 1000 // MB ->bytes

			if lang.CompileImage != "" {
				j.Compile = &ExecRequest{
					Image:          lang.CompileImage,
					Cmd:            lang.CompileCommand,
					SourceFileName: lang.SourceFileName,
				}
			}

			j.Exec = ExecRequest{
				Image:          lang.ExecImage,
				Cmd:            lang.ExecCommand,
				SourceFileName: "",
			}

			casesChan := make(chan TestCase, len(req.Cases))
			statusChan := make(chan JudgeStatus, 10)

			var check Judge
			var checkerCasesChan chan TestCase
			var checkerStatusChan chan JudgeStatus

			if req.Problem.Type == sctypes.JudgeRunningCode {
				checkerCodeBytes, err := ioutil.ReadAll(checker)

				if err != nil {
					pclient.JudgeSubmissionsUpdateResult(
						req.Submission.Cid,
						req.Submission.Sid,
						req.Submission.Jid,
						sctypes.SubmissionStatusInternalError,
						0, 0, 0,
						strings.NewReader("ReadAll(checker) error"),
					)

					return
				}
				var checkerInfo database.CheckerSavedFormat

				if err := json.Unmarshal(checkerCodeBytes, &checkerInfo); err != nil {
					pclient.JudgeSubmissionsUpdateResult(
						req.Submission.Cid,
						req.Submission.Sid,
						req.Submission.Jid,
						sctypes.SubmissionStatusInternalError,
						0, 0, 0,
						strings.NewReader("checker's format is illegal"),
					)

					return
				}

				check.Time = 3 * 1000
				check.Code = strings.NewReader(checkerInfo.Code)
				check.Mem = 256 * 1000 * 1000

				lmutex.RLock()
				lang, has := languages[checkerInfo.Lid]
				lmutex.RUnlock()

				if !has {
					pclient.JudgeSubmissionsUpdateResult(
						req.Submission.Cid,
						req.Submission.Sid,
						req.Submission.Jid,
						sctypes.SubmissionStatusInternalError,
						0, 0, 0,
						strings.NewReader("Unknown language(with lid:"+strconv.FormatInt(checkerInfo.Lid, 10)+") for checker"),
					)

					return
				}

				if lang.CompileImage != "" {
					check.Compile = &ExecRequest{
						Image:          lang.CompileImage,
						Cmd:            lang.CompileCommand,
						SourceFileName: lang.SourceFileName,
					}
				}

				check.Exec = ExecRequest{
					Image:          lang.ExecImage,
					Cmd:            lang.ExecCommand,
					SourceFileName: "",
				}

				checkerCasesChan = make(chan TestCase, len(req.Cases))
				checkerStatusChan = make(chan JudgeStatus, 10)
			}

			var lastResultResponded bool
			var releaser func()
			/*TODO:
			defer func() {
				if releaser != nil {
					releaser()
				}
			}
			*/
			go func() {
				defer close(casesChan)
				for i := range req.Problem.Cases {
					if err := ppd.Download(fs.FS_CATEGORY_TESTCASE_INOUT, req.Problem.Cases[i].Input); err != nil {
						pclient.JudgeSubmissionsUpdateResult(
							req.Submission.Cid,
							req.Submission.Sid,
							req.Submission.Jid,
							sctypes.SubmissionStatusInternalError,
							0, 0, 0,
							strings.NewReader("Download of testcases error("+req.Problem.Cases[i].Input+")"),
						)
						lastResultResponded = true
						return
					} else {
						locker, _ := ppd.NewLocker(fs.FS_CATEGORY_TESTCASE_INOUT, req.Problem.Cases[i].Input)

						releaser = utility.FunctionJoin(releaser, func() {
							locker.Unlock()
						})
					}

					if err := ppd.Download(fs.FS_CATEGORY_TESTCASE_INOUT, req.Problem.Cases[i].Input); err != nil {
						pclient.JudgeSubmissionsUpdateResult(
							req.Submission.Cid,
							req.Submission.Sid,
							req.Submission.Jid,
							sctypes.SubmissionStatusInternalError,
							0, 0, 0,
							strings.NewReader("Download of testcases error("+req.Problem.Cases[i].Input+")"),
						)
						lastResultResponded = true
						return
					} else {
						locker, _ := ppd.NewLocker(fs.FS_CATEGORY_TESTCASE_INOUT, req.Problem.Cases[i].Input)

						releaser = utility.FunctionJoin(releaser, func() {
							locker.Unlock()
						})
					}
				}
			}()

			go j.Run(statusChan, casesChan)

			go func() {
				totalStatus := sctypes.SubmissionStatusAccepted

				defer close(checkerCasesChan)

				for {
					stat, has := <-statusChan

					if !has {
						return
					}

					if stat.Case == -1 {
						totalStatus = stat.Status

						res := pclient.JudgeSubmissionsUpdateResult(
							req.Problem.Cid,
							req.Submission.Sid,
							req.Submission.Jid,
							totalStatus,
							0,
							stat.Time,
							stat.Mem,
							strings.NewReader(stat.Stderr),
						)

						// TODO: RunningCode
						return
					} else {
						pclient.JudgeSubmissionsUpdateCase(
							req.Problem.Cid,
							req.Submission.Sid,
							req.Submission.Jid,
							fmt.Sprintf("%d/%d", stat.Case, len(req.Problem.Cases)),
							database.SubmissionTestCase{
								Sid:    req.Submission.Sid,
								Cid:    req.Problem.Cid,
								Status: stat.Status,
								CaseID: stat.Case,
								Name:   req.Problem.Cases[stat.Case].Name,
								Time:   stat.Time,
								Mem:    stat.Mem,
							},
						)

						if status != transfer.Accepted {
							trans.ResponseChan <- res
						} else if req.Type == transfer.JudgeRunningCode {
							respArr[stat.Case] = res
							checkerCasesChan <- TCType{ID: stat.Case, In: "hogehoge"}
						} else {
							if stat.Stdout != req.Cases[stat.Case].Output {
								res.Status = transfer.WrongAnswer
								totalStatus = transfer.WrongAnswer
							}
							trans.ResponseChan <- res
						}
					}
				}
			}()

			if req.Type == transfer.JudgeRunningCode {
				go func() {
					setupFinished := false
					for {
						stat, has := <-checkerStatusChan

						if !has {
							return
						}

						if stat.Case == -1 {
							resp := respArr[len(respArr)-1]

							if stat.JR == Finished {
								if setupFinished {
									resp.Status = transfer.Accepted
								}
							} else if stat.JR == RuntimeError {
								resp.Status = transfer.WrongAnswer
							} else {
								resp.Msg = "Checker Program: " + JudgeResultCodeToStr[stat.JR]
								resp.Status = transfer.InternalError

								if stat.JR == CompileError {
									resp.Msg = stat.Stderr
								}
							}

							trans.ResponseChan <- resp

							return
						} else {
							setupFinished = true
							resp := respArr[stat.Case]

							if stat.JR == Finished {
								resp.Status = transfer.Accepted
							} else if stat.JR == RuntimeError {
								resp.Status = transfer.WrongAnswer
							} else {
								resp.Msg = "Checker Program: " + JudgeResultCodeToStr[stat.JR]
								resp.Status = transfer.InternalError
							}

							trans.ResponseChan <- resp
						}
					}
				}()
			}
		}()
	}

	/*exe, err := NewExecutor("Hello", 100*1024*1024, []string{"/host_tmp/a.out"}, "ubuntu:16.04", []string{"/tmp:/host_tmp:ro"}, "")

	if err != nil {
		fmt.Println(err)

		return
	}

	res := exe.Run(1000, "Hello")
	*/
	/*
		j := Judge{}

		j.Code = `
			#include <iostream>

			int main() {
				long long ll = 0;

				for(int i = 0; i < 100000000; ++i) {
					ll += i;
				}
				std::cout << "Hello, world" << std::endl;
			}
		`
		j.Compile = &ExecRequest{
			Cmd:            []string{"g++", "-std=c++14", "/work/main.cpp", "-o", "/work/a.out"},
			Image:          "ubuntu-mine:16.04",
			SourceFileName: "main.cpp",
		}
		j.Exec = ExecRequest{
			Cmd:            []string{"/work/a.out"},
			Image:          "ubuntu-mine:16.04",
			SourceFileName: "",
		}
		j.Mem = 100 * 1024 * 1024
		j.Time = 2000

		js := make(chan JudgeStatus, 10)
		tc := make(chan TCType, 10)

		go j.Run(js, tc)

		tc <- TCType{In: "", Name: "Test01"}
		close(tc)

		for c, res := <-js; res; c, res = <-js {
			var cas, stdout, stderr string
			if c.Stdout != nil {
				stdout = *c.Stdout
			} else {
				stdout = "<nil>"
			}
			if c.Stderr != nil {
				stderr = *c.Stderr
			} else {
				stderr = "<nil>"
			}
			if c.Case != nil {
				cas = *c.Case
			} else {
				cas = "<nil>"
			}
			fmt.Printf("Case: %s, Stdout: %s, Stderr: %s, Result: %s, Memory: %dKB, Time: %dms\n", cas, stdout, stderr, JudgeResultCodeToStr[int(c.JR)], c.Mem/1000, c.Time)
		}

		//	fmt.Println(res.ExitCode, res.Mem, res.Time, res.Status, res.Stdout, res.Stderr)

		//	err = exe.Delete()
		/*	err =
			if err != nil {
				fmt.Println(err)
			}

		fmt.Println(*wdir, *server)*/
}
