package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"log"
	"strconv"

	"context"

	transfer "github.com/cs3238-tsuzu/popcon-judge-go/Transfer"
	"github.com/cs3238-tsuzu/popcon-sc/ppjc/client"
	"github.com/cs3238-tsuzu/popcon-sc/ppjc/types"
	"github.com/docker/engine-api/client"
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
	wdir := flag.String("used-dir", "/tmp/pj", "Directory to execute programs and save files")
	server := flag.String("server", "http://192.168.2.1:8080/", "popcon-sc server address")
	parallel := flag.Int64("parallel", 1, "execution is parallel")
	auth := flag.String("auth", "", "authentication token")
	cgroup := flag.String("cgroup", "/sys/fs/cgroup", "cgroup dir")
	docker := flag.String("docker", "unix:///var/run/docker.sock", "docker host")

	flag.Parse()

	if help != nil && *help {
		flag.PrintDefaults()

		return
	}

	cgroupDir = *cgroup

	err := os.MkdirAll(*wdir, 0664)

	if err != nil {
		log.Println(err.Error())

		os.Exit(1)

		return
	}

	var languages map[int64]Language

	if true {
		fp, err := os.OpenFile(*settings, os.O_RDONLY, 0664)

		if err != nil {
			log.Println(err.Error())

			os.Exit(1)

			return
		}

		dec := json.NewDecoder(fp)

		err = dec.Decode(&settingData)

		if err != nil {
			log.Println("Failed to decode a json: " + err.Error())

			os.Exit(1)

			return
		}

		languages = make(map[int64]Language)

		for k, v := range settingData.Languages {
			lid, err := strconv.ParseInt(k, 10, 64)

			if err != nil {
				panic(err)
			}

			languages[lid] = v
		}
	}

	headers := map[string]string{"User-Agent": "popcon-judge/v1.00"}

	cli, err = client.NewClient(*docker, "v1.22", nil, headers)

	if err != nil {
		panic(err)
	}

	pclient, err := ppjc.NewClient(*server, *auth)

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

			lang, has := languages[req.Lang]

			if !has {
				trans.ResponseChan <- transfer.JudgeResponse{
					Sid:    req.Sid,
					Status: transfer.InternalError,
					Msg:    "Unknown Language",
					Case:   -1,
				}

				return
			}

			j.Code = req.Code
			j.Time = req.Time * 1000      // sec -> ms
			j.Mem = req.Mem * 1000 * 1000 // MB ->bytes

			if lang.Compile {
				j.Compile = &ExecRequest{
					Image:          lang.CompileImage,
					Cmd:            lang.CompileCmd,
					SourceFileName: lang.SourceFileName,
				}
			}

			j.Exec = ExecRequest{
				Image:          lang.ExecImage,
				Cmd:            lang.ExecCmd,
				SourceFileName: "",
			}

			casesChan := make(chan TCType, len(req.Cases))
			statusChan := make(chan JudgeStatus, 10)

			var check Judge
			var checkerCasesChan chan TCType
			var checkerStatusChan chan JudgeStatus

			if req.Type == transfer.JudgeRunningCode {
				check.Time = 3 * 1000
				check.Code = req.Checker
				check.Mem = 256 * 1000 * 1000

				lang, has := languages[req.CheckerLang]

				if !has {
					trans.ResponseChan <- transfer.JudgeResponse{
						Sid:    req.Sid,
						Status: transfer.InternalError,
						Msg:    "Unknown Language for Checker Program",
						Case:   -1,
					}

					return
				}

				if lang.Compile {
					check.Compile = &ExecRequest{
						Image:          lang.CompileImage,
						Cmd:            lang.CompileCmd,
						SourceFileName: lang.SourceFileName,
					}
				}

				check.Exec = ExecRequest{
					Image:          lang.ExecImage,
					Cmd:            lang.ExecCmd,
					SourceFileName: "",
				}

				checkerCasesChan = make(chan TCType, len(req.Cases))
				checkerStatusChan = make(chan JudgeStatus, 10)
			}

			go j.Run(statusChan, casesChan)

			go func() {
				for k, v := range req.Cases {
					casesChan <- TCType{ID: k, In: v.Input}
				}
				close(casesChan)
			}()

			respArr := make([]transfer.JudgeResponse, len(req.Cases)+1)

			respArr[len(respArr)-1] = transfer.JudgeResponse{
				Sid:    req.Sid,
				Status: transfer.InternalError,
				Case:   -1,
			}

			go func() {
				totalStatus := transfer.Accepted

				defer close(checkerCasesChan)

				for {
					stat, has := <-statusChan

					if !has {
						return
					}

					if stat.Case == -1 {
						if stat.JR == MemoryLimitExceeded {
							totalStatus = transfer.MemoryLimitExceeded
						} else if stat.JR == TimeLimitExceeded {
							totalStatus = transfer.MemoryLimitExceeded
						} else if stat.JR == RuntimeError {
							totalStatus = transfer.RuntimeError
						} else if stat.JR == InternalError {
							totalStatus = transfer.InternalError
						} else if stat.JR >= 6 && stat.JR <= 8 {
							totalStatus = transfer.CompileError

							if stat.JR == CompileTimeLimitExceeded {
								stat.Stderr = "Compile Time Limit Exceeded"
							} else if stat.JR == CompileMemoryLimitExceeded {
								stat.Stderr = "Compile Memory Limit Exceeded"
							}
						}

						res := transfer.JudgeResponse{
							Sid:    req.Sid,
							Status: totalStatus,
							Case:   -1,
							Msg:    stat.Stderr,
							Time:   stat.Time,
							Mem:    stat.Mem / 1000,
						}

						if req.Type == transfer.JudgePerfectMatch {
							trans.ResponseChan <- res
						} else {
							respArr[len(respArr)-1] = res
						}

						return
					} else {
						status := transfer.Accepted
						if stat.JR == Judging {
							trans.ResponseChan <- transfer.JudgeResponse{
								Sid:    req.Sid,
								Status: transfer.Judging,
								Case:   stat.Case,
								Msg:    fmt.Sprint(stat.Case, "/", len(req.Cases)),
								Time:   stat.Time,
								Mem:    stat.Mem / 1000,
							}
							continue
						}

						if stat.JR == MemoryLimitExceeded {
							status = transfer.MemoryLimitExceeded
						} else if stat.JR == TimeLimitExceeded {
							status = transfer.MemoryLimitExceeded
						} else if stat.JR == RuntimeError {
							status = transfer.RuntimeError
						} else if stat.JR == InternalError {
							status = transfer.InternalError
						}

						res := transfer.JudgeResponse{
							Sid:      req.Sid,
							Status:   status,
							Case:     stat.Case,
							CaseName: req.Cases[stat.Case].Name,
							Time:     stat.Time,
							Mem:      stat.Mem / 1000,
						}

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
