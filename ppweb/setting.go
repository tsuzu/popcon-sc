package main

import (
	"sync"
)

type Setting struct {
	CanCreateUser                    bool     `json:"can_create_user"`
	CanCreateContestByNotAdmin       bool     `json:"can_create_contest"`
	NumberOfDisplayedNews            int      `json:"number_of_news"`
	LogFile                          string   `json:"log_file"`
	CertificationWithEmail           bool     `json:"cert_with_mail"`
	SendmailCommand                  []string `json:"send_mail"`                  // ex. ["sendmail", "#{to}", "#{subject}", "#{body}"]
	CSRFTokenExpirationInMinutes     int64    `json:"csrf_token_expiration"`      // 分
	MailConfTokenExpirationInMinutes int64    `json:"mail_conf_token_expiration"` // 分
	MailMinimumIntervalInMinutes     int64    `json:"mail_min_interval"`          // 分
	SessionExpirationInMinutes       int      `json:"session_expiration"`         // 分
	PublicHost                       string   `json:"public_host"`                // ex. https://example.com

	// If HTTPS isn't needed, leave empty
	CertFilePath string `json:"cert_file"`
	KeyFilePath  string `json:"key_file"`

	// Environmental Variables
	dbAddr              string
	redisAddr           string
	redisPass           string
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
	// ロックする必要なし
	//	sm.mut.Lock()
	//	defer sm.mut.Unlock()

	sm.setting = setting
}

func (sm *SettingManager) Get() *Setting {
	// ロックする必要なし
	//	sm.mut.RLock()
	//	defer sm.mut.RUnlock()

	return &sm.setting
}

var settingManager SettingManager
