package redis

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/popcon-sc/lib/setting"
	"github.com/garyburd/redigo/redis"
)

type RedisManager struct {
	pool   *redis.Pool
	logger func() *logrus.Entry
	*ppconfiguration.RedisSettingManager
}

func NewRedisManager(addr, pass string, logger func() *logrus.Entry) (*RedisManager, error) {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr, redis.DialPassword(pass), redis.DialConnectTimeout(60*time.Second))
		},
		TestOnBorrow: func(c redis.Conn, _ time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	sm, err := ppconfiguration.NewRedisSettingManager(pool)

	if err != nil {
		return nil, err
	}

	return &RedisManager{
		pool:                pool,
		RedisSettingManager: sm,
		logger:              logger,
	}, nil
}

func (rm *RedisManager) TokenExists(service, token string) (bool, error) {
	conn := rm.pool.Get()
	defer conn.Close()
	key := "token_" + service + "_" + token + "_" + service

	v, err := redis.Int(conn.Do("EXISTS", key))

	if err != nil {
		return false, err
	}

	return v != 0, nil
}

func (rm *RedisManager) TokenRegister(service, token string, expiration time.Duration) error {
	return rm.TokenRegisterWithValue(service, token, expiration, 0)
}

func (rm *RedisManager) TokenRegisterWithValue(service, token string, expiration time.Duration, val interface{}) (err error) {
	conn := rm.pool.Get()
	defer conn.Close()

	key := "token_" + service + "_" + token + "_" + service

	if err := conn.Send("SETEX", key, strconv.FormatInt(int64(expiration.Seconds()), 10), val); err != nil {
		return err
	}

	return conn.Flush()
}

func (rm *RedisManager) TokenCheckAndRemove(service, token string) (bool, error) {
	conn := rm.pool.Get()
	defer conn.Close()
	key := "token_" + service + "_" + token + "_" + service

	cnt, err := redis.Int(conn.Do("DEL", key))

	if err != nil {
		return false, err
	}

	if cnt > 0 {
		return true, nil
	}

	return false, nil
}

func (rm *RedisManager) TokenGetAndRemove(service, token string) (bool, interface{}, error) {
	conn := rm.pool.Get()
	defer conn.Close()
	key := "token_" + service + "_" + token + "_" + service

	val, err := conn.Do("GET", key)

	if err != nil {
		if err == redis.ErrNil {
			return false, nil, nil
		}
		return false, nil, err
	}
	if _, err := conn.Do("DEL", key); err != nil {
		rm.logger().WithError(err).Error("Token deletion failed")
	}

	return true, val, nil
}

func (rm *RedisManager) TokenGetAndRemoveInt64(service, token string) (bool, int64, error) {
	r, v, e := rm.TokenGetAndRemove(service, token)

	if e != nil {
		return false, 0, e
	}

	if !r {
		return false, 0, nil
	}

	if ret, err := redis.Int64(v, nil); err != nil {
		if err == redis.ErrNil {
			return false, 0, nil
		}
		return false, 0, err
	} else {
		return true, ret, nil
	}
}

func (rm *RedisManager) TokenGetAndRemoveString(service, token string) (bool, string, error) {
	r, v, e := rm.TokenGetAndRemove(service, token)

	if e != nil {
		return false, "", e
	}

	if !r {
		return false, "", nil
	}

	if ret, err := redis.String(v, nil); err != nil {
		if err == redis.ErrNil {
			return false, "", nil
		}
		return false, "", fmt.Errorf("Type casting to string failed(val: %v)", v)
	} else {
		return true, ret, nil
	}
}

func (rm *RedisManager) TokenGenerate() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)

	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), err
}

func (rm *RedisManager) TokenGenerateAndRegister(service string, expiration time.Duration) (string, error) {
	return rm.TokenGenerateAndRegisterWithValue(service, expiration, 0)
}

func (rm *RedisManager) TokenGenerateAndRegisterWithValue(service string, expiration time.Duration, val interface{}) (string, error) {
	token, err := rm.TokenGenerate()

	if err != nil {
		return "", err
	}

	err = rm.TokenRegisterWithValue(service, token, expiration, val)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (rm *RedisManager) UniqueFileID(category string) (int64, error) {
	conn := rm.pool.Get()
	defer conn.Close()

	return redis.Int64(conn.Do("INCR", "unique_"+category+"_file_id_incr_"+category))
}

func (rm *RedisManager) JudgingProgressUpdate(sid int64, status string) error {
	conn := rm.pool.Get()
	defer conn.Close()

	if err := conn.Send("SET", "judging_progress_"+strconv.FormatInt(sid, 10), status); err != nil {
		return err
	}

	return conn.Flush()
}

func (rm *RedisManager) JudgingProgressGet(sid int64) (string, error) {
	conn := rm.pool.Get()
	defer conn.Close()

	return redis.String(conn.Do("GET", "judging_progress_"+strconv.FormatInt(sid, 10)))
}

func (rm *RedisManager) JudgingProgressDelete(sid int64) error {
	conn := rm.pool.Get()
	defer conn.Close()

	if err := conn.Send("DEL", "judging_progress_"+strconv.FormatInt(sid, 10)); err != nil {
		return err
	}

	return conn.Flush()
}

func (rm *RedisManager) Ping() error {
	conn := rm.pool.Get()
	defer conn.Close()

	pong, err := redis.String(conn.Do("PING"))

	if err != nil {
		return err
	} else if pong != "PONG" {
		return errors.New(pong)
	}

	return nil
}

func (rm *RedisManager) Close() {
	// Do nothing
}
