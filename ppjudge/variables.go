package main

import "github.com/docker/docker/client"

var workingDirectory string
var workingDirectoryHost string
var cli *client.Client
var CPUUtilization int // 0-100
