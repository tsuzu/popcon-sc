package main

type EnvironmentalSetting struct {
	// Environmental Variables
	dbAddr              string
	mongoAddr           string
	redisAddr           string
	redisPass           string
	judgeControllerAddr string
	microServicesAddr   string
	internalToken       string
	listeningEndpoint   string
	dataDirectory       string
	debugMode           bool
}

var environmentalSetting *EnvironmentalSetting
