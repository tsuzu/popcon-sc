package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
)

type JudgeSimple struct {
	Code    io.Reader
	Compile *ExecRequest
	Exec    ExecRequest
	Time    int64
	Mem     int64
}

func (j *JudgeSimple) Run(ch chan<- JudgeStatus, tests <-chan TestCase, replaceNewlineChar bool) {
	defer close(ch)

	// Identification
	id := RandomName()

	// Working Directory
	path := filepath.Join(workingDirectory, id)

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
		ch <- CreateInternalError(TotalResultCaseID, "Failed to create source file."+err.Error())

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

	// Compile
	if j.Compile != nil {
		exe, err := NewExecutor(id, 512*1024*1024, 10000, j.Compile.Cmd, j.Compile.Image, []string{filepath.Join(workingDirectoryHost, id) + ":/work"}, j.Compile.Env)

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

	exe, err := NewExecutor(id, j.Mem, j.Time, j.Exec.Cmd, j.Exec.Image, []string{filepath.Join(workingDirectoryHost, id) + ":" + "/work:ro"}, j.Exec.Env)

	if err != nil {
		ch <- CreateInternalError(-1, "Failed to create a Docker container to judge."+err.Error())

		return
	}

	defer exe.Delete()

	totalResult := sctypes.SubmissionStatusAccepted
	maxInt64 := func(a int64, b int64) int64 {
		if a > b {
			return a
		} else {
			return b
		}
	}

	var maxTime, maxMem int64

	for tc := range tests {
		GeneralLog().WithField("tc", tc).Debug("new testcase")
		name := tc.ID

		ch <- JudgeStatus{Case: name, Status: sctypes.SubmissionStatusJudging}

		r := sctypes.SubmissionStatusAccepted

		var inputStr string
		if b, err := ioutil.ReadFile(tc.Input); err != nil {
			ch <- CreateInternalError(name, "Failed to open a file(testcase input)")

			maxMem = -1
			maxTime = -1
			r = sctypes.SubmissionStatusInternalError
		} else {
			if replaceNewlineChar {
				inputStr = NewlineReplacer.Replace(string(b))
			} else {
				inputStr = string(b)
			}
			b = nil

			if res := exe.Run(inputStr); res.Status != ExecFinished {
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
					ch <- JudgeStatus{Case: name, Status: sctypes.SubmissionStatusAccepted, Mem: res.Mem, Time: res.Time, Stdout: res.Stdout, Stderr: res.Stderr}
					if maxMem != -1 {
						maxMem = maxInt64(maxMem, res.Mem)
					}
					if maxTime != -1 {
						maxTime = maxInt64(maxTime, res.Time)
					}
				}
			}
		}

		totalResult = maxStatus(totalResult, r)
	}

	ch <- JudgeStatus{Case: TotalResultCaseID, Status: totalResult, Time: maxTime, Mem: maxMem}
}

// Legacy
/*
	// User
	_, err := exec.Command("useradd", "--no-create-home", id).Output()

	if err != nil {
		ch <- CreateInternalError("Failed to create a directory to build your code. " + err.Error())

		return
	}

	uid, err := user.Lookup(id)

	if err != nil {
		ch <- CreateInternalError("Failed to look up a user. " + err.Error())

		return
	}

	uidInt, err := strconv.ParseInt(uid.Uid, 10, 64)
	if err != nil {
		ch <- CreateInternalError("Failed to parseInt uid. " + err.Error())

		return
	}

	gidInt, err := strconv.ParseInt(uid.Gid, 10, 64)
	if err != nil {
		ch <- CreateInternalError("Failed to parseInt gid. " + err.Error())

		return
	}

	defer exec.Command("userdel", id).Output()
*/
