package main

import (
	"sync"
)

type Setting struct {
	CanCreateUser              bool   `json:"can_create_user"`
	CanCreateContestByNotAdmin bool   `json:"can_create_contest"`
	NumberOfDisplayedNews      int    `json:"number_of_news"`
	LogFile                    string `json:"log_file"`

	// If HTTPS isn't needed, leave empty
	CertFilePath string `json:"cert_file"`
	KeyFilePath  string `json:"key_file"`

	dbAddr              string
	judgeControllerAddr string
	internalToken       string
	listeningEndpoint   string
	dataDirectory       string
}

type SettingManager struct {
	setting Setting
	mut     sync.RWMutex
}

func (sm *SettingManager) Set(setting Setting) {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	sm.setting = setting
}

func (sm *SettingManager) Get() Setting {
	sm.mut.RLock()
	defer sm.mut.RUnlock()

	return sm.setting
}

var settingManager SettingManager
