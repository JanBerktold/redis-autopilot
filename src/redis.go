//go:generate go-enum -f=$GOFILE
package main

import (
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-redis/redis"
)

// RedisStatus is an enumeration of all possible states the health of a redis instance can have.
/*
ENUM(
Ready
Loading
Syncing
NotResponding
Unknown
)
*/
type RedisStatus int

type RedisInstance interface {
	Status() RedisStatus
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

func (r *redisInstance) Status() RedisStatus {
	if r.client.Ping().Err != nil {
		return RedisStatusNotResponding
	}

	// Check to get instance info
	rawInfo, err := r.client.Info().Result()

	if err != nil {
		return RedisStatusUnknown
	}

	info := parseRedisInfo(rawInfo)

	// Check for ongoing loading from existing rdb or aof backup.
	if loading, ok := info["loading"]; ok && loading == "1" {
		return RedisStatusLoading
	} else if !ok {
		return RedisStatusUnknown
	}

	// Check for ongoing SYNC from a master.
	if _, ok := info["master_sync_left_bytes"]; ok {
		return RedisStatusSyncing
	}

	return RedisStatusReady
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

type RedisWatcher interface {
	Status() RedisStatus
	ChangeChannel() <-chan RedisStatus
}
