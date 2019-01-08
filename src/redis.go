package main

import (
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-redis/redis"
)

type RedisInstance interface {
	MakeSlave(masterAddr net.TCPAddr) error
	MakeMaster() error
}

type redisInstance struct {
	client *redis.Client
}

func NewRedisInstance(addr string) RedisInstance {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DialTimeout:  1 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	})

	return &redisInstance{
		client: client,
	}
}

func (r *redisInstance) MakeSlave(masterAddr net.TCPAddr) error {
	return r.client.SlaveOf(masterAddr.IP.String(), strconv.Itoa(masterAddr.Port)).Err()
}

func (r *redisInstance) MakeMaster() error {
	return r.client.SlaveOf("NO", "ONE").Err()
}

func parseRedisInfo(in string) map[string]string {
	out := make(map[string]string)
	lines := strings.Split(in, "\r\n")
	for _, line := range lines {
		trimmed := strings.TrimFunc(line, unicode.IsSpace)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.Split(trimmed, ":")

		if len(parts) < 2 {
			continue
		}

		out[parts[0]] = parts[1]
	}
	return out
}
