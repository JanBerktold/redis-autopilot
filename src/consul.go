package main

import (
	consul "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"time"
)

type ConsulClient interface {
	RegisterService(name string, ttl time.Duration) error
}

type consulClient struct {
	client    *consul.Client
	aclGetter func() string
}

func NewConsulClient(address string, aclGetter func() string) (ConsulClient, error) {
	client, err := consul.NewClient(&consul.Config{
		Address: address,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create consul client")
	}

	return &consulClient{
		client:    client,
		aclGetter: aclGetter,
	}, nil
}

func (c *consulClient) RegisterService(name string, ttl time.Duration) error {
	serviceDef := &consul.AgentServiceRegistration{
		Name: name,
		Check: &consul.AgentServiceCheck{
			TTL: ttl.String(),
		},
	}

	return c.client.Agent().ServiceRegister(serviceDef)
}

type ConsulClientProvider interface {
	Instance() ConsulClient
	ChangeChannel() <-chan ConsulClient
}

type consulClientProvider struct {
	instance        ConsulClient
	consulURLGetter func() string
	consulACLGetter func() string
	changeSignal    <-chan struct{}
	listeners       []chan ConsulClient
	logger          *log.Logger
}

func NewConsulClientProvider(consulURLGetter func() string, consulACLGetter func() string, changeSignal <-chan struct{}, logger *log.Logger) (ConsulClientProvider, error) {
	if consulURLGetter == nil {
		return nil, errors.New("consulURLGetter is nil")
	}
	if consulACLGetter == nil {
		return nil, errors.New("consulACLGetter is nil")
	}
	if changeSignal == nil {
		return nil, errors.New("changeSignal is nil")
	}
	if logger == nil {
		return nil, errors.New("logger is nil")
	}

	provider := &consulClientProvider{
		consulURLGetter: consulURLGetter,
		consulACLGetter: consulACLGetter,
		changeSignal:    changeSignal,
		listeners:       make([]chan ConsulClient, 0),
		logger:          logger,
	}

	go provider.monitor()

	instance, err := NewConsulClient(consulURLGetter(), consulACLGetter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create initial consul client")
	}
	provider.instance = instance

	return provider, nil
}

func (c *consulClientProvider) Instance() ConsulClient {
	return c.instance
}

func (c *consulClientProvider) ChangeChannel() <-chan ConsulClient {
	channel := make(chan ConsulClient)
	c.listeners = append(c.listeners, channel)
	return channel
}

func (c *consulClientProvider) monitor() {
	for {
		<-c.changeSignal

		addr := c.consulURLGetter()
		instance, err := NewConsulClient(addr, c.consulACLGetter)
		if err != nil {
			c.logger.Errorf("Failed to create new consul client for %q: %v", addr, err)
			continue
		}

		c.logger.Infof("Refreshed consul client, now pointing at: %v", addr)
		c.instance = instance

		for _, listener := range c.listeners {
			select {
			case listener <- instance:
			default:
			}
		}
	}
}

type ConsulServiceRegistrar interface {
	Enable() error
	Disable() error
	Enabled() bool
}

type consulServiceRegistrar struct {
	enabled              bool
	consulClientProvider ConsulClientProvider
	serviceNameGetter    func() string
	ttlGetter            func() time.Duration
	changeSignal         <-chan struct{}
}

func NewConsulServiceRegistrar(consulClientProvider ConsulClientProvider, serviceNameGetter func() string,
	ttlGetter func() time.Duration, changeSignal <-chan struct{}) (ConsulServiceRegistrar, error) {

	registrar := &consulServiceRegistrar{
		enabled:              true,
		consulClientProvider: consulClientProvider,
		serviceNameGetter:    serviceNameGetter,
		ttlGetter:            ttlGetter,
		changeSignal:         changeSignal,
	}

	if err := registrar.register(); err != nil {
		return nil, err
	}

	go registrar.continuouslyUpdateTTL()

	return registrar, nil
}

func (r *consulServiceRegistrar) Enable() error {
	r.enabled = true
	return nil
}

func (r *consulServiceRegistrar) Disable() error {
	r.enabled = false
	return nil
}

func (r *consulServiceRegistrar) Enabled() bool {
	return r.enabled
}

func (r *consulServiceRegistrar) register() error {
	instance := r.consulClientProvider.Instance()
	serviceName := r.serviceNameGetter()
	ttl := r.ttlGetter()
	return instance.RegisterService(serviceName, ttl)
}

func (r *consulServiceRegistrar) continuouslyUpdateTTL() {

}
