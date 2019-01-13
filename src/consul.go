package main

import (
	consul "github.com/hashicorp/consul/api"
)

type ConsulClient interface {
}

type consulClient struct {
	client *consul.Client
}

type ConsulClientProvider interface {
}

type ConsulServiceRegistrar interface {
	Enable() error
	Disable() error
	Enabled() bool
}

type consulServiceRegistrar struct {
	enabled bool
}

func NewConsulServiceRegistrar() (ConsulServiceRegistrar, error) {
	return &consulServiceRegistrar{}, nil
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
