//go:generate go-enum -f=$GOFILE
package main

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
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
	logger *log.Logger
}

func NewRedisInstance(addr string, logger *log.Logger) RedisInstance {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  1 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	})

	return &redisInstance{
		client: client,
		logger: logger,
	}
}

func (r *redisInstance) Status() RedisStatus {
	if err := r.client.Ping().Err(); err != nil {
		r.logger.Infof("failed to ping redis instance: %v\n", err)
		return RedisStatusNotResponding
	}

	// Check to get instance info
	rawInfo, err := r.client.Info().Result()
	if err != nil {
		r.logger.Infof("failed to retrieve INFO from redis instance: %v\n", err.Error())
		return RedisStatusUnknown
	}

	info := parseRedisInfo(rawInfo)

	// Check for ongoing loading from existing rdb or aof backup.
	if loading, ok := info["loading"]; ok && loading == "1" {
		return RedisStatusLoading
	} else if !ok {
		r.logger.Infof("INFO result did not include loading key\n")
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

type redisWatcher struct {
	redisInstanceProvider RedisInstanceProvider
	monitorIntervalGetter func() time.Duration
	lastStatus            RedisStatus
	listeners             []chan RedisStatus
	logger                *log.Logger
}

func NewRedisWatcher(instanceProvider RedisInstanceProvider, logger *log.Logger, monitorIntervalGetter func() time.Duration) (RedisWatcher, error) {
	watcher := &redisWatcher{
		redisInstanceProvider: instanceProvider,
		monitorIntervalGetter: monitorIntervalGetter,
		listeners:             make([]chan RedisStatus, 0),
		logger:                logger,
	}

	watcher.update()
	go watcher.monitor()

	return watcher, nil
}

func (r *redisWatcher) Status() RedisStatus {
	return r.lastStatus
}

func (r *redisWatcher) ChangeChannel() <-chan RedisStatus {
	return make(chan RedisStatus)
}

func (r *redisWatcher) monitor() {
	for {
		time.Sleep(r.monitorIntervalGetter())
		r.update()
	}
}

func (r *redisWatcher) update() {
	status := r.redisInstanceProvider.Instance().Status()

	if r.lastStatus == status {
		r.logger.Info("Performed redis health check with no change. Status: %v", status.String())
		return
	}

	r.lastStatus = status

	for _, listener := range r.listeners {
		select {
		case listener <- status:
		default:
		}
	}
}

type RedisInstanceProvider interface {
	Instance() RedisInstance
	ChangeSignal() <-chan RedisInstance
}

type redisInstanceProvider struct {
	currentInstance RedisInstance
	changeSignal    <-chan struct{}
	redisAddrGetter func() string
	listeners       []chan RedisInstance
	logger          *log.Logger
}

func NewRedisInstanceProvider(redisAddrGetter func() string, changeSignal <-chan struct{}, logger *log.Logger) (RedisInstanceProvider, error) {
	if redisAddrGetter == nil {
		return nil, errors.New("redisAddrGetter is nil")
	}
	if changeSignal == nil {
		return nil, errors.New("changeSignal is nil")
	}
	if logger == nil {
		return nil, errors.New("logger is nil")
	}

	instance := NewRedisInstance(redisAddrGetter(), logger)

	provider := &redisInstanceProvider{
		currentInstance: instance,
		changeSignal:    changeSignal,
		redisAddrGetter: redisAddrGetter,
		listeners:       make([]chan RedisInstance, 0),
		logger:          logger,
	}

	go provider.monitor()

	return provider, nil
}

func (p *redisInstanceProvider) Instance() RedisInstance {
	return p.currentInstance
}

func (p *redisInstanceProvider) ChangeSignal() <-chan RedisInstance {
	signal := make(chan RedisInstance)
	p.listeners = append(p.listeners, signal)
	return signal
}

func (p *redisInstanceProvider) monitor() {
	for {
		<-p.changeSignal

		redisAddr := p.redisAddrGetter()
		instance := NewRedisInstance(redisAddr, p.logger)
		p.currentInstance = instance

		p.logger.Infof("Refreshed redis instance with %v\n.", redisAddr)

		for _, listener := range p.listeners {
			select {
			case listener <- instance:
			default:
			}
		}
	}
}
