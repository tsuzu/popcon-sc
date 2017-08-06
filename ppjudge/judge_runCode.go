package main

import (
	"io"
	"os"
	"path/filepath"
	"io/ioutil"
	"fmt"
	"strings"
	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
)

type JudgeRunCode struct {
	JudgeSimple
	CheckerCode    io.Reader
	CheckerCompile *ExecRequest
	CheckerExec    ExecRequest
}

func (j *JudgeRunCode) Run(ch chan<- JudgeStatus, tests <-chan TestCase) {
	defer close(ch)

	// Identification
	id := RandomName()
	// Identification for Checker
	idc := RandomName() // Checker is ran in this directory.
	idj := RandomName() // To store input files

	// Working Directory for Checker
	cpath := filepath.Join(workingDirectory, idc)
	jpath := filepath.Join(workingDirectory, idj)

	if true {
		// jpath
		err := os.Mkdir(jpath, 0777)

		if err != nil {
			ch <- CreateInternalError(TotalResultCaseID, "Failed to create a directory for checker. "+err.Error())

			return
		}

		defer os.RemoveAll(jpath)

		err = os.Chmod(jpath, 0777)

		if err != nil {
			ch <- CreateInternalError(TotalResultCaseID, "Failed to chmod the directory for checker. "+err.Error())

			return
		}

		// cpath
		err = os.Mkdir(cpath, 0777)

		if err != nil {
			ch <- CreateInternalError(TotalResultCaseID, "Failed to create a directory for checker. "+err.Error())

			return
		}

		defer os.RemoveAll(cpath)

		err = os.Chmod(cpath, 0777)

		if err != nil {
			ch <- CreateInternalError(TotalResultCaseID, "Failed to chmod the directory for checker. "+err.Error())

			return
		}

		// Source File
		fp, err := os.Create(filepath.Join(cpath, j.CheckerExec.SourceFileName))

		if err != nil {
			ch <- CreateInternalError(TotalResultCaseID, "Failed to create source code file for checker. "+err.Error())

			return
		}

		_, err = io.Copy(fp, j.CheckerCode)

		if err != nil {
			ch <- CreateInternalError(TotalResultCaseID, "Failed to write your code on your file for checker. "+err.Error())

			return
		}

		fp.Close()

		err = os.Chmod(filepath.Join(cpath, j.CheckerExec.SourceFileName), 0644)

		if err != nil {
			ch <- CreateInternalError(TotalResultCaseID, "Failed to chmod the source file for checker. "+err.Error())

			return
		}
	}

	// Working Directory
	path := filepath.Join(workingDirectory, id)

	// Judged Code
	if true {
		err := os.Mkdir(path, 0777)

		if err != nil {
			ch <- CreateInternalError(TotalResultCaseID, "Failed to create a directory. "+err.Error())

			return
		}

		defer os.RemoveAll(path)

		err = os.Chmod(path, 0777)

		if err != nil {
			ch <- CreateInternalError(TotalResultCaseID, "Failed to chmod the directory. "+err.Error())

			return
		}

		// Source File
		fp, err := os.Create(filepath.Join(path, j.Exec.SourceFileName))

		if err != nil {
			ch <- CreateInternalError(TotalResultCaseID, "Failed to create source code file."+err.Error())

			return
		}

		_, err = io.Copy(fp, j.Code)

		if err != nil {
			ch <- CreateInternalError(TotalResultCaseID, "Failed to write your code on your file. "+err.Error())

			return
		}

		fp.Close()

		err = os.Chmod(filepath.Join(path, j.Exec.SourceFileName), 0644)

		if err != nil {
			ch <- CreateInternalError(TotalResultCaseID, "Failed to chmod the source file. "+err.Error())

			return
		}
	}

	// Compile for Checker
	if j.CheckerCompile != nil {
		exe, err := NewExecutor(idc, 512*1024*1024, 10000, j.CheckerCompile.Cmd, j.CheckerCompile.Image, []string{filepath.Join(workingDirectoryHost, idc) + ":/work"}, j.CheckerCompile.Env)

		if err != nil {
			ch <- CreateInternalError(-1, "Failed to create a Docker container to compile checker code."+err.Error())

			return
		}

		res := exe.Run("")

		exe.Delete()
		if res.Status != ExecFinished {
			switch res.Status {
			case ExecError:
				ch <- CreateInternalError(TotalResultCaseID, "Failed to execute a compiler for checker. "+res.Stderr)

				return
			case ExecMemoryLimitExceeded:
				ch <- CreateInternalError(TotalResultCaseID, "Memory Limit Exceeded(checker code)")

				return
			case ExecTimeLimitExceeded:
				ch <- CreateInternalError(TotalResultCaseID, "Time Limit Exceeded(checker code)")

				return
			}
		}

		if res.ExitCode != 0 {
			ch <- CreateInternalError(TotalResultCaseID, "Checker Code Compile Error(message is hidden)")

			return
		}
	}

	// Compile
	if j.Compile != nil {
		exe, err := NewExecutor(id, 512*1024*1024, 10000, j.Compile.Cmd, j.Compile.Image, []string{filepath.Join(workingDirectoryHost, id) + ":" + "/work"}, j.Compile.Env)

		if err != nil {
			ch <- CreateInternalError(-1, "Failed to create a Docker container to compile your code."+err.Error())

			return
		}

		res := exe.Run("")

		exe.Delete()
		if res.Status != ExecFinished {
			switch res.Status {
			case ExecError:
				ch <- CreateInternalError(TotalResultCaseID, "Failed to execute a compiler. "+res.Stderr)

				return
			case ExecMemoryLimitExceeded:
				ch <- JudgeStatus{Case: TotalResultCaseID, Status: sctypes.SubmissionStatusCompileError, Stderr: "Memory Limit Exceeded"}

				return
			case ExecTimeLimitExceeded:
				ch <- JudgeStatus{Case: TotalResultCaseID, Status: sctypes.SubmissionStatusCompileError, Stderr: "Time Limit Exceeded"}

				return
			}
		}

		if res.ExitCode != 0 {
			ch <- JudgeStatus{Case: TotalResultCaseID, Status: sctypes.SubmissionStatusCompileError, Stderr: res.Stderr}

			return
		}
	}

	maxInt64 := func(a int64, b int64) int64 {
		if a > b {
			return a
		} else {
			return b
		}
	}

	exe, err := NewExecutor(id, j.Mem, j.Time, j.Exec.Cmd, j.Exec.Image, []string{filepath.Join(workingDirectoryHost, id) + ":/work:ro"}, j.Exec.Env)

	if err != nil {
		ch <- CreateInternalError(-1, "Failed to create a Docker container to judge."+err.Error())

		return
	}

	defer exe.Delete()

	cexe, err := NewExecutor(idc, 512*1024*1024, 10, j.CheckerExec.Cmd, j.CheckerExec.Image, []string{filepath.Join(workingDirectoryHost, idc) + ":/work:ro", filepath.Join(workingDirectoryHost, idj) + ":/data:ro", filepath.Join(workingDirectoryHost, id) + ":/judged:ro"}, j.CheckerExec.Env)

	defer cexe.Delete()

	var maxTime, maxMem int64

	totalResult := sctypes.SubmissionStatusAccepted

	for tc := range tests {
		name := tc.ID

		ch <- JudgeStatus{Case: name, Status: sctypes.SubmissionStatusJudging}

		r := sctypes.SubmissionStatusAccepted
		res := exe.Run(tc.Input)

		if res.Status != ExecFinished {
			switch res.Status {
			case ExecError:
				ch <- CreateInternalError(name, "Failed to execute your code. "+res.Stderr)
				r = sctypes.SubmissionStatusInternalError
				maxMem = -1
				maxTime = -1
			case ExecMemoryLimitExceeded:
				ch <- JudgeStatus{Case: name, Status: sctypes.SubmissionStatusMemoryLimitExceeded}
				r = sctypes.SubmissionStatusMemoryLimitExceeded
				maxMem = -1
				maxTime = -1
			case ExecTimeLimitExceeded:
				ch <- JudgeStatus{Case: name, Status: sctypes.SubmissionStatusTimeLimitExceeded}
				r = sctypes.SubmissionStatusTimeLimitExceeded
				maxMem = -1
				maxTime = -1
			}
		} else {
			if res.ExitCode != 0 {
				ch <- JudgeStatus{Case: name, Status: sctypes.SubmissionStatusRuntimeError}
				r = sctypes.SubmissionStatusRuntimeError
				maxMem = -1
				maxTime = -1
			} else {
				if err := ioutil.WriteFile(filepath.Join(jpath, "stdin.txt"), []byte(tc.Input), 0444); err != nil {
					ch <- CreateInternalError(name, "Failed to write stdin in stdin.txt.:"+err.Error())
					r = sctypes.SubmissionStatusInternalError
					maxMem = -1
					maxTime = -1
				} else if err := ioutil.WriteFile(filepath.Join(jpath, "stdout.txt"), []byte(res.Stdout), 0444); err != nil {
					ch <- CreateInternalError(name, "Failed to write stdout in stdout.txt.:"+err.Error())
					r = sctypes.SubmissionStatusInternalError
					maxMem = -1
					maxTime = -1
				} else {
					// Release memory
					tc.Input = ""

					cres := cexe.Run(fmt.Sprintf("%d\n%s\n%s\n%s", name, "/data/stdin.txt", "/data/stdout.txt", filepath.Join("/judged", j.Exec.SourceFileName)))

					if cres.Status != ExecFinished {
						switch cres.Status {
						case ExecError:
							ch <- CreateInternalError(name, "Failed to execute checker code.")
							r = sctypes.SubmissionStatusInternalError
							maxMem = -1
							maxTime = -1
						case ExecMemoryLimitExceeded:
							ch <- CreateInternalError(name, "Memory limit of the checker exceeded.")
							r = sctypes.SubmissionStatusInternalError
							maxMem = -1
							maxTime = -1
						case ExecTimeLimitExceeded:
							ch <- CreateInternalError(name, "Time limit of the checker exceeded.")
							r = sctypes.SubmissionStatusInternalError
							maxMem = -1
							maxTime = -1
						}
					} else {
						if cres.ExitCode != 0 {
							ch <- CreateInternalError(name, "Runtime of the checker error")
							r = sctypes.SubmissionStatusInternalError
							maxMem = -1
							maxTime = -1
						} else {
							// TODO: Support flexible score
							// s := strings.SplitN(res.Stdout, "\n", 2)
							// s = strings.SplitN(s[0], "\r", 2)

							if strings.SplitN(strings.SplitN(cres.Stdout, "\n", 2)[0], "\r", 2)[0] == "WA" {
								r = sctypes.SubmissionStatusWrongAnswer
							}
							// strconv.ParseInt(s[0], 10, 64)
							ch <- JudgeStatus{
								Case:   name,
								Status: r,
								Mem:    res.Mem,
								Time:   res.Time,
								Stdout: res.Stdout,
								Stderr: res.Stderr,
							}
						}
					}
				}

				if maxMem != -1 {
					maxMem = maxInt64(maxMem, res.Mem)
				}
				if maxTime != -1 {
					maxTime = maxInt64(maxTime, res.Time)
				}
			}
		}

		totalResult = maxStatus(totalResult, r)
	}

	ch <- JudgeStatus{Case: TotalResultCaseID, Status: totalResult, Time: maxTime, Mem: maxMem}

}
