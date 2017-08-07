package main

import (
	"archive/tar"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/net/context"
)

var CgroupParentPrefix = "/" // Will be exchanged if running on Docker

type Executor struct {
	Name string
	Mem  int64
	Time int64
	Cgr  Cgroup
}

type ExecStatus int

const (
	ExecFinished            ExecStatus = 0
	ExecTimeLimitExceeded   ExecStatus = 1
	ExecMemoryLimitExceeded ExecStatus = 2
	ExecError               ExecStatus = 3
)

type ExecResult struct {
	Status   ExecStatus
	Time     int64
	Mem      int64
	ExitCode int
	Stdout   string
	Stderr   string
}

func (e *Executor) Run(input string) ExecResult {
	cg := e.Cgr
	memc := cg.getSubsys("memory")

	hijack, err := cli.ContainerAttach(context.Background(), e.Name, types.ContainerAttachOptions{Stream: true, Stdin: true, Stdout: true, Stderr: true})

	if err != nil {
		return ExecResult{ExecError, 0, 0, 0, "", "Failed to hijack container: " + err.Error()}
	}

	defer hijack.Close()

	if err := os.MkdirAll(filepath.Join(workingDirectory, "stdouterr"), 0777); err != nil {
		return ExecResult{ExecError, 0, 0, 0, "", "Failed to mkdir: " + err.Error()}
	}

	stdout, err := ioutil.TempFile(filepath.Join(workingDirectory, "stdouterr"), "stdout")
	if err != nil {
		return ExecResult{ExecError, 0, 0, 0, "", "Failed to create a temporary file: " + err.Error()}
	}
	defer func() {
		stdout.Close()
		os.Remove(stdout.Name())
	}()

	stderr, err := ioutil.TempFile(filepath.Join(workingDirectory, "stdouterr"), "stderr")
	if err != nil {
		return ExecResult{ExecError, 0, 0, 0, "", "Failed to create a temporary file: " + err.Error()}
	}
	defer func() {
		stderr.Close()
		os.Remove(stderr.Name())
	}()

	var hijackErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := hijack.Conn.Write([]byte(input))

		if err != nil {
			hijackErr = err

			return
		}

		hijack.CloseWrite()

		_, err = stdcopy.StdCopy(stdout, stderr, hijack.Reader)

		if err != nil {
			hijackErr = err

			return
		}
	}()

	ctx := context.Background()
	err = cli.ContainerStart(ctx, e.Name, types.ContainerStartOptions{})

	if err != nil {
		return ExecResult{ExecError, 0, 0, 0, "", "Failed to start a container. " + err.Error()}
	}

	wg.Wait()

	if hijackErr != nil {
		return ExecResult{ExecError, 0, 0, 0, "", "Failed to hijack stdin/stdout"}
	}

	const LimitedSize int64 = 100 * 1024 * 1024
	var stdoutStr, stderrStr string
	func() {
		if _, e := stdout.Seek(0, 0); err != nil {
			err = e
			return
		}
		if _, e := stderr.Seek(0, 0); err != nil {
			err = e
			return
		}

		b, e := ioutil.ReadAll(&io.LimitedReader{stdout, LimitedSize})

		if err != nil {
			err = e
			return
		}
		stdoutStr = string(b)

		b, err = ioutil.ReadAll(&io.LimitedReader{stderr, LimitedSize})

		if err != nil {
			err = e
			return
		}
		stderrStr = string(b)
	}()

	if err != nil {
		return ExecResult{ExecError, 0, 0, 0, "", "Failed to read stdout/stderr: " + err.Error()}
	}

	rc, _, err := cli.CopyFromContainer(ctx, e.Name, "/tmp/time.txt")

	if err != nil {
		cli.ContainerKill(ctx, e.Name, "SIGKILL")

		return ExecResult{ExecError, 0, 0, 0, "", "Failed to read the execution time. " + err.Error()}
	}

	tarStream := tar.NewReader(rc)
	tarStream.Next()

	buf, err := ioutil.ReadAll(tarStream)

	if err != nil {
		cli.ContainerKill(context.Background(), e.Name, "SIGKILL")

		return ExecResult{ExecError, 0, 0, 0, "", "Failed to read tar stream: " + err.Error()}
	}

	var exitCode int
	var execMillisec int64
	for _, elm := range strings.Split(strings.TrimRight(NewlineReplacer.Replace(string(buf)), "\n"), "\n") {
		arrRes := strings.Split(elm, " ")

		if len(arrRes) != 2 {
			cli.ContainerKill(ctx, e.Name, "SIGKILL")

			if a := regexp.MustCompilePOSIX("[tT]erminated by signal ([0-9]*)$").FindStringSubmatch(elm); len(a) != 0 {
				s, _ := strconv.ParseInt(a[1], 10, 64)

				exitCode = 128 + int(s)

				continue
			} else {
				return ExecResult{ExecError, 0, 0, 0, "", "Failed to parse the result." + elm}
			}
		}

		execSec, err := strconv.ParseFloat(arrRes[0], 64)

		if err != nil {
			return ExecResult{ExecError, 0, 0, 0, "", "Failed to parse the execution time."}
		}

		execMillisec = int64(execSec * 1000)

		exit, err := strconv.ParseInt(strings.Split(arrRes[1], "\n")[0], 10, 32)

		if err != nil {
			return ExecResult{ExecError, 0, 0, 0, "", "Failed to parse the exit code."}
		}

		if exitCode == 0 {
			exitCode = int(exit)
		}
	}

	if execMillisec > e.Time {
		cli.ContainerKill(ctx, e.Name, "SIGKILL")

		return ExecResult{ExecTimeLimitExceeded, 0, 0, 0, "", ""}
	}

	usedMem, err := memc.getValInt("memory.max_usage_in_bytes")

	if usedMem >= e.Mem {
		return ExecResult{ExecMemoryLimitExceeded, 0, 0, 0, "", ""}
	}

	return ExecResult{ExecFinished, execMillisec, usedMem, exitCode, stdoutStr, stderrStr}
}

func (e *Executor) Delete() error {
	err := e.Cgr.Delete()
	err2 := cli.ContainerRemove(context.Background(), e.Name, types.ContainerRemoveOptions{Force: true})

	errstr := ""
	if err != nil {
		errstr += err.Error()
	}
	if err2 != nil {
		errstr += err2.Error()
	}

	if errstr == "" {
		return nil
	} else {
		return errors.New(errstr)
	}
}

func NewExecutor(name string, mem int64, time int64, cmd []string, img string, binds []string, env []string) (*Executor, error) {
	ctx := context.Background()

	cg := NewCgroup(name)

	err := cg.addSubsys("memory").Modify()

	if err != nil {
		return nil, errors.New("Failed to create a cgroup")
	}

	err = cg.getSubsys("memory").setValInt(mem, "memory.limit_in_bytes")

	if err != nil {
		cg.Delete()

		return nil, errors.New("Failed to set memory.limit_in_bytes")
	}

	// Usage of swapping should not be restricted.
	/*err = cg.getSubsys("memory").setValInt(mem, "memory.memsw.limit_in_bytes")

	if err != nil {
		cg.Delete()

		return nil, errors.New("Failed to set memory.memsw.limit_in_bytes")
	}*/

	cfg := container.Config{}

	cfg.Tty = false
	cfg.AttachStderr = true
	cfg.AttachStdout = true
	cfg.AttachStdin = true
	cfg.OpenStdin = true
	cfg.StdinOnce = true
	cfg.Image = img
	cfg.Env = env
	cfg.Hostname = "localhost"

	var timer = []string{"/usr/bin/time", "-q", "-f", "%e %x", "-o", "/tmp/time.txt", "/usr/bin/timeout", strconv.FormatInt((time+999)/1000, 10), "/usr/bin/sudo", "-u", "nobody", "-E"}

	newCmd := make([]string, 0, len(cmd)+len(timer))

	for i := range timer {
		newCmd = append(newCmd, timer[i])
	}

	for i := range cmd {
		newCmd = append(newCmd, cmd[i])
	}

	cfg.Cmd = newCmd

	hcfg := container.HostConfig{}

	hcfg.CPUQuota = int64(1000 * CPUUtilization)
	hcfg.CPUPeriod = 100000
	hcfg.NetworkMode = "none"
	hcfg.Binds = binds
	hcfg.CgroupParent = "/" + name

	_, err = cli.ContainerCreate(ctx, &cfg, &hcfg, nil, name)

	if err != nil {
		cg.Delete()

		return nil, errors.New("Failed to create a container " + err.Error())
	}

	return &Executor{name, mem, time, cg}, nil
}
