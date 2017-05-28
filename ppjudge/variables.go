package main

import "github.com/docker/engine-api/client"

var settingData SettingsInterface
var workingDirectory string
var cli *client.Client
var CPUUtilization int // 0-100
