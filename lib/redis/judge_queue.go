package redis

import (
	"context"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
)

const JudgeQueueName = "popcon_judge_queue"

type JudgeQueueQuery string

func (jqq JudgeQueueQuery) Parse() (int64, int64) {
	arr := strings.Split(string(jqq), ":")

	if len(arr) != 2 {
		return -1, -1
	}

	s, _ := strconv.ParseInt(arr[0], 10, 64)
	t, _ := strconv.ParseInt(arr[1], 10, 64)

	return s, t
}

func (jqq *JudgeQueueQuery) Format(cid, sid int64) {
	*jqq = JudgeQueueQuery(strconv.FormatInt(cid, 10) + ":" + strconv.FormatInt(sid, 10))
}

func NewJudgeQueueQuery(cid, sid int64) JudgeQueueQuery {
	var jqq JudgeQueueQuery
	jqq.Format(cid, sid)

	return jqq
}

func (rm *RedisManager) JudgeQueuePush(cid, sid int64) error {
	conn := rm.pool.Get()

	_, err := conn.Do("LPUSH", JudgeQueueName, string(NewJudgeQueueQuery(cid, sid)))

	return err
}

func (rm *RedisManager) JudgeQueuePopBlockingWithContext(timeout int64, ctx context.Context) (int64, int64, error) {
	conn := rm.pool.Get()

	for {
		select {
		case <-ctx.Done():
			return 0, 0, ctx.Err()
		default:
		}
		res, err := conn.Do("BRPOP", JudgeQueueName, timeout)

		if err != nil {
			if err == redis.ErrNil {
				continue
			}
			return 0, 0, err
		}
		if res == nil {
			continue
		}

		if arr, ok := res.([]interface{}); ok && len(arr) == 2 {
			if b, ok := arr[1].([]byte); ok {
				jqq := JudgeQueueQuery(b)
				cid, sid := jqq.Parse()
				return cid, sid, nil
			} else {
				return 0, 0, ErrIllegalFormat
			}
		} else {
			return 0, 0, ErrIllegalFormat
		}
	}
}

func (rm *RedisManager) JudgeQueuePopBlocking(timeout int64) (int64, int64, error) {
	return rm.JudgeQueuePopBlockingWithContext(timeout, context.Background())
}

func (rm *RedisManager) JudgeQueuePop() (int64, int64, error) {
	conn := rm.pool.Get()

	res, err := redis.String(conn.Do("RPOP", JudgeQueueName))

	if err != nil {
		return 0, 0, err
	}

	jqq := JudgeQueueQuery(res)

	cid, sid := jqq.Parse()

	return cid, sid, nil
}

func (rm *RedisManager) JudgeIDGet() (int64, error) {
	conn := rm.pool.Get()

	return redis.Int64(conn.Do("INCR", "popcon_sc_jid"))
}
