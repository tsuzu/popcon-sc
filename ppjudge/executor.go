package main

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
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

	stdinErr := make(chan error, 1)
	stdoutErr := make(chan error, 1)
	stderrErr := make(chan error, 1)

	attachment := func(opt types.ContainerAttachOptions, done chan<- error, out *string) {
		ctx := context.Background()
		hijack, err := cli.ContainerAttach(ctx, e.Name, opt)

		if err != nil {
			panic(err)
		}
		done <- nil
		if opt.Stdin {
			hijack.Conn.Write([]byte(input))
			hijack.CloseWrite()
			hijack.Close()
			hijack.Conn.Close()

			done <- nil
			return
		}

		var buf bytes.Buffer
		for {
			b := make([]byte, 128)

			size, err := hijack.Reader.Read(b)

			// Output Size Limitation
			if out != nil && len(*out) < 100*1024*1024 {
				buf.Write(b[0:size])

				if buf.Len() >= 8 {
					var size uint32
					bin := buf.Bytes()

					for i, v := range bin[4:8] {
						shift := uint32((3 - i) * 8)

						size |= uint32(v) << shift
					}

					if buf.Len() >= int(size+8) {
						*out += string(bin[8 : size+8])
						buf.Reset()
						buf.Write(bin[size+8:])
					}
				}
			}

			if err != nil {
				if err == io.EOF {
					done <- nil
				} else {
					done <- err
				}

				return
			}
		}
	}

	var stdout, stderr string

	go attachment(types.ContainerAttachOptions{Stream: true, Stdout: true}, stdoutErr, &stdout)
	go attachment(types.ContainerAttachOptions{Stream: true, Stderr: true}, stderrErr, &stderr)
	go attachment(types.ContainerAttachOptions{Stream: true, Stdin: true}, stdinErr, nil)

	<-stdinErr
	<-stdoutErr
	<-stderrErr
	<-stdinErr

	ctx := context.Background()
	err := cli.ContainerStart(ctx, e.Name, types.ContainerStartOptions{})

	if err != nil {
		return ExecResult{ExecError, 0, 0, 0, "", "Failed to start a container. " + err.Error()}
	}

	<-stdoutErr
	<-stderrErr

	rc, _, err := cli.CopyFromContainer(ctx, e.Name, "/tmp/time.txt")

	if err != nil {
		cli.ContainerKill(ctx, e.Name, "SIGKILL")

		return ExecResult{ExecError, 0, 0, 0, "", "Failed to read the execution time. " + err.Error()}
	}

	tarStream := tar.NewReader(rc)
	tarStream.Next()

	buf := new(bytes.Buffer)
	buf.ReadFrom(tarStream)
	arrRes := strings.Split(buf.String(), " ")

	if len(arrRes) != 2 {
		cli.ContainerKill(ctx, e.Name, "SIGKILL")

		return ExecResult{ExecError, 0, 0, 0, "", "Failed to parse the result."}
	}

	execSec, err := strconv.ParseFloat(arrRes[0], 64)

	if err != nil {
		return ExecResult{ExecError, 0, 0, 0, "", "Failed to parse the execution result."}
	}

	execTime := int64(execSec * 1000)

	exit64, err := strconv.ParseInt(strings.Split(arrRes[1], "\n")[0], 10, 32)

	if err != nil {
		return ExecResult{ExecError, 0, 0, 0, "", "Failed to parse the exit code."}
	}

	exitCode := int(exit64)

	if execTime > e.Time {
		cli.ContainerKill(ctx, e.Name, "SIGKILL")

		return ExecResult{ExecTimeLimitExceeded, 0, 0, 0, "", ""}
	}

	usedMem, err := memc.getValInt("memory.max_usage_in_bytes")

	if usedMem >= e.Mem {
		return ExecResult{ExecMemoryLimitExceeded, 0, 0, 0, "", ""}
	}

	return ExecResult{ExecFinished, execTime, usedMem, exitCode, stdout, stderr}
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
