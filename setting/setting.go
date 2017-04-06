package ppconfiguration

import (
	"strings"

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

type Structure struct {
	CanCreateUser           bool
	CanCreateContest        bool
	NumberOfDisplayedNews   int
	CertificationWithEmail  bool
	SendMailCommand         []string
	CSRFConfTokenExpiration int64
	MailConfTokenExpiration int64
	MailMinInterval         int64
	SessionExpiration       int
	StandardSignupGroup     int64
	PublicHost              string
}

var setting = map[SettingNameType]ConfigTypeOfValue{
	CanCreateUser:           ConfigTypeBool,
	CanCreateContest:        ConfigTypeBool,
	NumberOfDisplayedNews:   ConfigTypeInt,
	CertificationWithEmail:  ConfigTypeBool,
	SendMailCommand:         ConfigTypeString, // []string -> split by ,
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
	CSRFConfTokenExpiration: int64(30),   // min
	MailConfTokenExpiration: int64(1400), // min
	MailMinInterval:         int64(10),   // min
	SessionExpiration:       1576800,     // min
	StandardSignupGroup:     int64(1),
	PublicHost:              "https://localhost",
}

func settingKeyName(settingName SettingNameType) string {
	return "setting_" + string(settingName)
}

type RedisSettingManager struct {
	pool *redis.Pool
}

func NewRedisSettingManager(pool *redis.Pool) (*RedisSettingManager, error) {
	conn := pool.Get()
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
	conn := rsm.pool.Get()
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
	conn := rsm.pool.Get()
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

func (rsm *RedisSettingManager) sendMailCommandParse(str string) []string {
	return strings.Split(str, ",")
}

func (rsm *RedisSettingManager) SendMailCommand() ([]string, error) {
	s, err := rsm.GetString(SendMailCommand)

	if err != nil {
		return nil, err
	}

	return rsm.sendMailCommandParse(s), nil
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

func (rsm *RedisSettingManager) GetAll() (*Structure, error) {
	var err error
	var str Structure
	conn := rsm.pool.Get()
	defer conn.Close()

	getter := func(s SettingNameType) {
		conn.Send("GET", settingKeyName(s))
	}
	getter(CanCreateUser)
	getter(CanCreateContest)
	getter(NumberOfDisplayedNews)
	getter(CertificationWithEmail)
	getter(SendMailCommand)
	getter(CSRFConfTokenExpiration)
	getter(MailConfTokenExpiration)
	getter(MailMinInterval)
	getter(SessionExpiration)
	getter(StandardSignupGroup)
	getter(PublicHost)
	if err := conn.Flush(); err != nil {
		return nil, err
	}

	str.CanCreateUser, err = redis.Bool(conn.Receive())
	if err != nil {
		return nil, err
	}
	str.CanCreateContest, err = redis.Bool(conn.Receive())
	if err != nil {
		return nil, err
	}
	str.NumberOfDisplayedNews, err = redis.Int(conn.Receive())
	if err != nil {
		return nil, err
	}
	str.CertificationWithEmail, err = redis.Bool(conn.Receive())
	if err != nil {
		return nil, err
	}

	s, err := redis.String(conn.Receive())
	str.SendMailCommand = rsm.sendMailCommandParse(s)
	if err != nil {
		return nil, err
	}

	str.CSRFConfTokenExpiration, err = redis.Int64(conn.Receive())
	if err != nil {
		return nil, err
	}
	str.MailConfTokenExpiration, err = redis.Int64(conn.Receive())
	if err != nil {
		return nil, err
	}
	str.MailMinInterval, err = redis.Int64(conn.Receive())
	if err != nil {
		return nil, err
	}
	str.SessionExpiration, err = redis.Int(conn.Receive())
	if err != nil {
		return nil, err
	}
	str.StandardSignupGroup, err = redis.Int64(conn.Receive())
	if err != nil {
		return nil, err
	}
	str.PublicHost, err = redis.String(conn.Receive())
	if err != nil {
		return nil, err
	}

	return &str, nil
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
	return rsm.Set(SendMailCommand, strings.Join(a, ","))
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

func (rsm *RedisSettingManager) SetAll(str *Structure) error {
	conn := rsm.pool.Get()

	setter := func(k SettingNameType, v interface{}) {
		conn.Send("SET", settingKeyName(k), v)
	}

	setter(CanCreateUser, str.CanCreateUser)
	setter(CanCreateContest, str.CanCreateContest)
	setter(NumberOfDisplayedNews, str.NumberOfDisplayedNews)
	setter(CertificationWithEmail, str.CertificationWithEmail)
	setter(SendMailCommand, strings.Join(str.SendMailCommand, ","))
	setter(CSRFConfTokenExpiration, str.CSRFConfTokenExpiration)
	setter(MailConfTokenExpiration, str.MailConfTokenExpiration)
	setter(MailMinInterval, str.MailMinInterval)
	setter(SessionExpiration, str.SessionExpiration)
	setter(StandardSignupGroup, str.StandardSignupGroup)
	setter(PublicHost, str.PublicHost)

	return conn.Flush()
}
