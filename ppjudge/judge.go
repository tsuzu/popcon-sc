package main

import ( //import "os/exec"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/cs3238-tsuzu/popcon-sc/lib/types"
	"github.com/seehuhn/mt19937"
)

//import "os/user"

type TestCase struct {
	ID    int64
	Input string
}

type ExecRequest struct {
	Image          string
	Cmd            []string
	SourceFileName string
}

type Judge struct {
	Code    io.Reader
	Compile *ExecRequest
	Exec    ExecRequest
	Time    int64
	Mem     int64
}

type JudgeStatus struct {
	Case   int64                        `json:"case"`
	Status sctypes.SubmissionStatusType `json:"jr"`
	Mem    int64                        `json:"mem"`
	Time   int64                        `json:"time"`
	Stdout string                       `json:"stdout"`
	Stderr string                       `json:"stderr"` // error and messageMsg
}

const TotalResultCaseID = -1

func CreateInternalError(id int64, msg string) JudgeStatus {
	return JudgeStatus{Case: id, Status: sctypes.SubmissionStatusInternalError, Stderr: msg}
}

const BASE_RAND_STRING = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomName() string {
	rng := rand.New(mt19937.New())
	rng.Seed(time.Now().UnixNano())

	res := make([]byte, 0, 16)
	for i := 0; i < 16; i++ {
		res = append(res, BASE_RAND_STRING[rng.Intn(len(BASE_RAND_STRING))])
	}

	return string(res)
}

func (j *Judge) Run(ch chan<- JudgeStatus, tests <-chan TestCase) {
	defer close(ch)

	// Identity
	id := RandomName()

	// Working Directory
	path := filepath.Join(workingDirectory, id)

	err := os.Mkdir(path, 0777)

	if err != nil {
		ch <- CreateInternalError(TotalResultCaseID, "Failed to create a directory. "+err.Error())

		return
	}

	defer os.RemoveAll(path)

	if err != nil {
		ch <- CreateInternalError(TotalResultCaseID, "Failed to chown the directory. "+err.Error())

		return
	}

	err = os.Chmod(path, 0777)

	if err != nil {
		ch <- CreateInternalError(TotalResultCaseID, "Failed to chmod the directory. "+err.Error())

		return
	}

	// Source File
	fp, err := os.Create(path + "/" + j.Compile.SourceFileName)

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

	err = os.Chmod(path+"/"+j.Compile.SourceFileName, 0644)

	if err != nil {
		ch <- CreateInternalError(TotalResultCaseID, "Failed to chmod the source file. "+err.Error())

		return
	}

	// Compile
	if j.Compile != nil {
		exe, err := NewExecutor(id, 512*1024*1024, 10000, j.Compile.Cmd, j.Compile.Image, []string{path + ":" + "/work"})

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

	exe, err := NewExecutor(id, j.Mem, j.Time, j.Exec.Cmd, j.Exec.Image, []string{path + ":" + "/work:ro"})

	if err != nil {
		ch <- CreateInternalError(-1, "Failed to create a Docker container to judge."+err.Error())

		return
	}

	defer exe.Delete()

	totalResult := sctypes.SubmissionStatusAccepted
	maxStatus := func(a, b sctypes.SubmissionStatusType) sctypes.SubmissionStatusType {
		_a := int(a)
		_b := int(b)
		if _a > _b {
			return a
		} else {
			return b
		}
	}

	maxInt := func(a int, b int) int {
		if a > b {
			return a
		} else {
			return b
		}
	}
	maxInt64 := func(a int64, b int64) int64 {
		if a > b {
			return a
		} else {
			return b
		}
	}

	var maxTime, maxMem int64

	for tc, res := <-tests; res; tc, res = <-tests {
		name := tc.ID

		ch <- JudgeStatus{Case: name, Status: sctypes.SubmissionStatusJudging}

		r := sctypes.SubmissionStatusAccepted
		res := exe.Run(tc.Input)

		if res.Status != ExecFinished {
			switch res.Status {
			case ExecError:
				msg := "Failed to execute your code. " + res.Stderr
				ch <- CreateInternalError(name, msg)
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
