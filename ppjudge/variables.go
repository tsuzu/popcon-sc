package main

import (
	"github.com/docker/docker/client"
	"strings"
)

var workingDirectory string
var workingDirectoryHost string
var cli *client.Client
var CPUUtilization int // 0-100

var NewlineReplacer = strings.NewReplacer("\r\n", "\n", "\r", "\n")
