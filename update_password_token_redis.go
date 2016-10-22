package main

import "github.com/garyburd/redigo/redis"

type SessionManager struct {
    pool *redis.Pool
}

func NewSessionManager(creator func(int)*redis.Pool) (*SessionManager, error) {
    return &SessionManager{creator(0)}, nil
}