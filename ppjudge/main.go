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

	"io/ioutil"

	"encoding/json"

	"regexp"

	"errors"

	"github.com/cs3238-tsuzu/popcon-sc/lib/database"
	"github.com/cs3238-tsuzu/popcon-sc/lib/filesystem"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/cs3238-tsuzu/popcon-sc/lib/utility"
	"github.com/k0kubun/pp"
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

func main() {
	cgroup := os.Getenv("PPJUDGE_CGROUP")             //flag.String("cgroup", "/sys/fs/cgroup", "cgroup dir")
	docker := os.Getenv("PPJUDGE_DOCKER")             //flag.String("docker", "unix:///var/run/docker.sock", "docker host path")
	dockerVer := os.Getenv("PPJUDGE_DOCKER_VER")      //flag.String("docker-ver", "v1.29", "version of docker api")
	onDocker := os.Getenv("PPJUDGE_ON_DOCKER") == "1" //flag.Bool("on-docker", false, "whether this is running on Docker")

	help := flag.Bool("help", false, "Display all options")
	wdir := flag.String("exec-dir", "/tmp/pj", "Directory to execute programs and save files")
	server := flag.String("server", "http://192.168.2.1:8080/", "popcon-sc server address for ppjc node")
	auth := flag.String("auth", "", "authentication token for popcon-sc ppjc")
	parallel := flag.Int64("parallel", 1, "the number of executions in parallel")
	cpuUsage := flag.Int("cpu-util", 100, "restriction of CPU utilization(expressed with per-cent, so must be 1-100)")
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
	CPUUtilization = *cpuUsage

	if onDocker {
		b, err := ioutil.ReadFile("/proc/self/cgroup")

		if err != nil {
			FSLog().WithError(err).Fatal("Failed to read /proc/self/cgroup")
		}

		if a := regexp.MustCompilePOSIX("(/docker/.*)$").FindStringSubmatch(string(b)); len(a) == 0 {
			FSLog().WithError(errors.New("cgroup not found")).Fatal("Failed to prepare cgroup environemnt")
		} else {
			CgroupParentPrefix = a[1] + "/"
		}
	}

	cgroupDir = cgroup

	// Cgroup Test
	if true {
		cg := NewCgroup("test_cgroup")

		if err := cg.addSubsys("memory").Modify(); err != nil {
			GeneralLog().WithError(err).Fatal("Failed to create a cgroup")
		}

		if err := cg.getSubsys("memory").setValInt(100*1024*1024, "memory.limit_in_bytes"); err != nil {
			GeneralLog().WithError(err).Fatal("Failed to setup a cgroup")
		}

		if err := cg.Delete(); err != nil {
			GeneralLog().WithError(err).Fatal("Failed to delete a cgroup")
		}
	}

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

	cli, err = client.NewClient(docker, dockerVer, nil, headers)

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
	reloadLanguagesCtx, reloadLanguagesCtxCanceller := context.WithCancel(context.Background())
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
			case <-reloadLanguagesCtx.Done():
				break
			}
		}
	}()

	pclient, err := ppjc.NewClient(*server, *auth)

	if err != nil {
		panic(err)
	}

	ppd, err := NewPPDownloader(pclient, *wdir)

	ppd.RunAutomaticallyDeleter(context.Background(), 30*time.Minute /*TODO: can change by prgoram options*/)

	if err != nil {
		panic(err)
	}

	jinfoChan := make(chan ppjctypes.JudgeInformation, 100)
	ctx, canceller := context.WithCancel(context.Background())

	defer canceller()

	_, oneFinish, err := pclient.StartWorkersWSPolling(*parallel, jinfoChan, ctx)

	if err != nil {
		panic(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-signalChan
		reloadLanguagesCtxCanceller()
		canceller()
	}()

	for {
		req, ok := <-jinfoChan

		if !ok {
			return
		}

		go func() {
			GeneralLog().WithField("req", pp.Sprint(req)).Debug("A new judge has started!")

			defer oneFinish()

			// TODO: Support interactive
			if req.Problem.Type == sctypes.JudgeInteractive {
				pclient.JudgeSubmissionsUpdateResult(
					req.Submission.Cid,
					req.Submission.Sid,
					req.Submission.Jid,
					sctypes.SubmissionStatusInternalError,
					0, 0, 0,
					strings.NewReader("Unsupported judge type"),
				)

				return
			}

			j := JudgeRunCode{}

			codePath, codeLocker, err := ppd.Download(fs.FS_CATEGORY_SUBMISSION, req.Submission.CodeFile)

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
			defer codeLocker.Unlock()

			var checkerPath string
			var checkerLocker *PPLocker
			if req.Problem.Type == sctypes.JudgeInteractive || req.Problem.Type == sctypes.JudgeRunningCode {
				checkerPath, checkerLocker, err = ppd.Download(fs.FS_CATEGORY_PROBLEM_CHECKER, req.Problem.CheckerFile)

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
			defer checkerLocker.Unlock()

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

			code, err := os.Open(codePath)

			if err != nil {
				pclient.JudgeSubmissionsUpdateResult(
					req.Submission.Cid,
					req.Submission.Sid,
					req.Submission.Jid,
					sctypes.SubmissionStatusInternalError,
					0, 0, 0,
					strings.NewReader("open source code files error"),
				)

				return
			}
			defer code.Close()

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
				SourceFileName: lang.SourceFileName,
			}

			casesChan := make(chan TestCase, len(req.Cases))
			statusChan := make(chan JudgeStatus, 10)

			if req.Problem.Type == sctypes.JudgeRunningCode {
				checkerCodeBytes, err := ioutil.ReadFile(checkerPath)

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

				j.CheckerCode = strings.NewReader(checkerInfo.Code)

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
					j.CheckerCompile = &ExecRequest{
						Image:          lang.CompileImage,
						Cmd:            lang.CompileCommand,
						SourceFileName: lang.SourceFileName,
					}
				}

				j.CheckerExec = ExecRequest{
					Image:          lang.ExecImage,
					Cmd:            lang.ExecCommand,
					SourceFileName: lang.SourceFileName,
				}
			}

			var judge Judge
			switch req.Problem.Type {
			case sctypes.JudgePerfectMatch:
				judge = &j.JudgeSimple
			case sctypes.JudgeRunningCode:
				judge = &j
			}

			var lastResultResponded bool
			var releaser func()

			var releaserWG sync.WaitGroup
			defer func() {
				releaserWG.Wait()
				if releaser != nil {
					releaser()
				}
			}()

			go func() {
				defer releaserWG.Done()
				defer close(casesChan)
				for i := range req.Problem.Cases {
					var inputPath string
					if path, locker, err := ppd.Download(fs.FS_CATEGORY_TESTCASE_INOUT, req.Problem.Cases[i].Input); err != nil {
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
						inputPath = path
						releaser = utility.FunctionJoin(releaser, func() {
							locker.Unlock()
						})
					}

					if _, locker, err := ppd.Download(fs.FS_CATEGORY_TESTCASE_INOUT, req.Problem.Cases[i].Output); err != nil {
						pclient.JudgeSubmissionsUpdateResult(
							req.Submission.Cid,
							req.Submission.Sid,
							req.Submission.Jid,
							sctypes.SubmissionStatusInternalError,
							0, 0, 0,
							strings.NewReader("Download of testcases error("+req.Problem.Cases[i].Output+")"),
						)
						lastResultResponded = true
						return
					} else {
						releaser = utility.FunctionJoin(releaser, func() {
							locker.Unlock()
						})
					}
					casesChan <- TestCase{
						ID:    int64(i),
						Input: inputPath,
					}
				}
			}()

			go judge.Run(statusChan, casesChan)

			go func() {
				totalStatus := sctypes.SubmissionStatusAccepted

				for {
					stat, has := <-statusChan

					if !has {
						return
					}

					if stat.Case == -1 {
						totalStatus = maxStatus(totalStatus, stat.Status)

						if err := pclient.JudgeSubmissionsUpdateResult(
							req.Problem.Cid,
							req.Submission.Sid,
							req.Submission.Jid,
							totalStatus,
							0,
							stat.Time,
							stat.Mem,
							strings.NewReader(stat.Stderr),
						); err != nil {
							HttpLog().WithError(err).Error("JudgeSubmissionsUpdateResult() error")
						}

						return
					} else {
						if stat.Status != sctypes.SubmissionStatusAccepted {
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
						} else if req.Problem.Type == sctypes.JudgePerfectMatch {
							fp, err := ppd.OpenFile(fs.FS_CATEGORY_TESTCASE_INOUT, req.Cases[stat.Case].Output)

							if err != nil {
								stat.Status = sctypes.SubmissionStatusInternalError

								FSLog().WithError(err).Error("File open error")
							} else if b, err := ioutil.ReadAll(fp); err != nil {
								stat.Status = sctypes.SubmissionStatusInternalError

								FSLog().WithError(err).Error("File readall error")
							} else if stat.Stdout != string(b) {
								stat.Status = sctypes.SubmissionStatusWrongAnswer
							}
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
						}
					}
				}
			}()
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
