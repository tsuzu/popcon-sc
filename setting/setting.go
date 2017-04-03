package ppconfiguration

import (
	"encoding/json"

	"github.com/garyburd/redigo/redis"
)

type ConfigTypeOfValue int

const (
	ConfigTypeInt ConfigTypeOfValue = iota
	ConfigTypeInt64
	ConfigTypeBool
	ConfigTypeString
)

type SettingNameType string

const (
	CanCreateUser           SettingNameType = "can_create_user"
	CanCreateContest        SettingNameType = "can_create_contest"
	NumberOfDisplayedNews   SettingNameType = "number_of_displayed_news"
	CertificationWithEmail  SettingNameType = "certification_with_email"
	SendMailCommand         SettingNameType = "sendmail_command"
	CSRFConfTokenExpiration SettingNameType = "cstf_token_expiration"
	MailConfTokenExpiration SettingNameType = "mail_conf_token_expiration"
	MailMinInterval         SettingNameType = "mail_min_interval"
	SessionExpiration       SettingNameType = "session_expiration"
	StandardSignupGroup     SettingNameType = "standard_signup_group"
	PublicHost              SettingNameType = "public_host"
)

var setting = map[SettingNameType]ConfigTypeOfValue{
	CanCreateUser:           ConfigTypeBool,
	CanCreateContest:        ConfigTypeBool,
	NumberOfDisplayedNews:   ConfigTypeInt,
	CertificationWithEmail:  ConfigTypeBool,
	SendMailCommand:         ConfigTypeString, // []string -> json
	CSRFConfTokenExpiration: ConfigTypeInt64,  // min
	MailConfTokenExpiration: ConfigTypeInt64,  // min
	MailMinInterval:         ConfigTypeInt64,  // min
	SessionExpiration:       ConfigTypeInt,    // min
	StandardSignupGroup:     ConfigTypeInt64,
	PublicHost:              ConfigTypeString,
}

var settingDefault = map[SettingNameType]interface{}{
	CanCreateUser:           false,
	CanCreateContest:        false,
	NumberOfDisplayedNews:   5,
	CertificationWithEmail:  true,
	SendMailCommand:         "",
	CSRFConfTokenExpiration: 30,      // min
	MailConfTokenExpiration: 1400,    // min
	MailMinInterval:         10,      // min
	SessionExpiration:       1576800, // min
	StandardSignupGroup:     1,
	PublicHost:              "https://localhost",
}

func settingKeyName(settingName SettingNameType) string {
	return "setting_" + string(settingName)
}

type RedisSettingManager struct {
	pool *redis.Pool
}

func NewRedisSettingManager(pool *redis.Pool) (*RedisSettingManager, error) {
	conn, err := pool.Dial()

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	for k, v := range settingDefault {
		conn.Send("SETNX", settingKeyName(k), v)
	}

	if err := conn.Flush(); err != nil {
		return nil, err
	}

	return &RedisSettingManager{
		pool: pool,
	}, nil
}

func (rsm *RedisSettingManager) Get(settingName SettingNameType) (interface{}, error) {
	conn, err := rsm.pool.Dial()

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.Do("GET", settingKeyName(settingName))
}

func (rsm *RedisSettingManager) GetString(settingName SettingNameType) (string, error) {
	return redis.String(rsm.Get(settingName))
}
func (rsm *RedisSettingManager) GetInt(settingName SettingNameType) (int, error) {
	return redis.Int(rsm.Get(settingName))
}
func (rsm *RedisSettingManager) GetInt64(settingName SettingNameType) (int64, error) {
	return redis.Int64(rsm.Get(settingName))
}
func (rsm *RedisSettingManager) GetBool(settingName SettingNameType) (bool, error) {
	return redis.Bool(rsm.Get(settingName))
}

func (rsm *RedisSettingManager) Set(settingName SettingNameType, val interface{}) error {
	conn, err := rsm.pool.Dial()

	if err != nil {
		return err
	}
	defer conn.Close()

	conn.Send("SET", settingKeyName(settingName), val)

	return conn.Flush()
}

func (rsm *RedisSettingManager) CanCreateUser() (bool, error) {
	return rsm.GetBool(CanCreateUser)
}
func (rsm *RedisSettingManager) CanCreateContest() (bool, error) {
	return rsm.GetBool(CanCreateContest)
}
func (rsm *RedisSettingManager) NumberOfDisplayedNews() (int, error) {
	return rsm.GetInt(NumberOfDisplayedNews)
}
func (rsm *RedisSettingManager) CertificationWithEmail() (bool, error) {
	return rsm.GetBool(CertificationWithEmail)
}
func (rsm *RedisSettingManager) SendMailCommand() ([]string, error) {
	s, err := rsm.GetString(SendMailCommand)

	if err != nil {
		return nil, err
	}

	var ret []string
	json.Unmarshal([]byte(s), ret)
	return ret, nil
}
func (rsm *RedisSettingManager) CSRFConfTokenExpiration() (int64, error) {
	return rsm.GetInt64(CSRFConfTokenExpiration)
}
func (rsm *RedisSettingManager) MailConfTokenExpiration() (int64, error) {
	return rsm.GetInt64(MailConfTokenExpiration)
}
func (rsm *RedisSettingManager) MailMinInterval() (int64, error) {
	return rsm.GetInt64(MailMinInterval)
}
func (rsm *RedisSettingManager) SessionExpiration() (int, error) {
	return rsm.GetInt(SessionExpiration)
}
func (rsm *RedisSettingManager) StandardSignupGroup() (int64, error) {
	return rsm.GetInt64(StandardSignupGroup)
}
func (rsm *RedisSettingManager) PublicHost() (string, error) {
	return rsm.GetString(PublicHost)
}

func (rsm *RedisSettingManager) CanCreateUserSet(a bool) error {
	return rsm.Set(CanCreateUser, a)
}
func (rsm *RedisSettingManager) CanCreateContestSet(a bool) error {
	return rsm.Set(CanCreateContest, a)
}
func (rsm *RedisSettingManager) NumberOfDisplayedNewsSet(a int) error {
	return rsm.Set(NumberOfDisplayedNews, a)
}
func (rsm *RedisSettingManager) CertificationWithEmailSet(a bool) error {
	return rsm.Set(CertificationWithEmail, a)
}
func (rsm *RedisSettingManager) SendMailCommandSet(a []string) error {
	b, _ := json.Marshal(a)

	return rsm.Set(SendMailCommand, string(b))
}
func (rsm *RedisSettingManager) CSRFConfTokenExpirationSet(a int64) error {
	return rsm.Set(CSRFConfTokenExpiration, a)
}
func (rsm *RedisSettingManager) MailConfTokenExpirationSet(a int64) error {
	return rsm.Set(MailConfTokenExpiration, a)
}
func (rsm *RedisSettingManager) MailMinIntervalSet(a int64) error {
	return rsm.Set(MailMinInterval, a)
}
func (rsm *RedisSettingManager) SessionExpirationSet(a int) error {
	return rsm.Set(SessionExpiration, a)
}
func (rsm *RedisSettingManager) StandardSignupGroupSet(a int64) error {
	return rsm.Set(StandardSignupGroup, a)
}
func (rsm *RedisSettingManager) PublicHostSet(a string) error {
	return rsm.Set(PublicHost, a)
}
