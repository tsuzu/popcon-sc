package main

import "github.com/garyburd/redigo/redis"

type RedisManager struct {
    conn *redis.Pool
}

func NewRedisPoolCreator(addr, pass string) func(int) *redis.Pool {
    return func(db int) *redis.Pool {
        return &redis.Pool{
            MaxIdle: 120,
            MaxActive: 12000,
            Dial: func() (redis.Conn, error) {
                c, err := redis.Dial("tcp", addr)
            
                if err != nil {
                    return nil, err
                }
            
                if len(pass) != 0 {
                    if _, err := c.Do("AUTH", pass); err != nil {
                        c.Close()
                        return nil, err
                    }        
                }

                if _, err := c.Do("SELECT", db); err != nil {
                    return nil, err
                }

                return c, err
            },
        }
    }
}