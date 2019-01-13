package main

type Pilot interface {
	Execute() error
	Shutdown()
}

type singleMasterWithSlavesPilot struct {
	redisInstanceProvider  RedisInstanceProvider
	consulServiceRegistrar ConsulServiceRegistrar
}

func NewSingleMasterWithSlavesPilot(redisInstanceProvider RedisInstanceProvider, consulServiceRegistrar ConsulServiceRegistrar) Pilot {
	return &singleMasterWithSlavesPilot{
		redisInstanceProvider:  redisInstanceProvider,
		consulServiceRegistrar: consulServiceRegistrar,
	}
}

func (p *singleMasterWithSlavesPilot) Execute() error {
	instance := p.redisInstanceProvider.Instance()

	redisStatus := instance.Status()
	if redisStatus != RedisStatusReady {
		// make sure we're not in consul
	}

	return nil
}

func (p *singleMasterWithSlavesPilot) Shutdown() {

}
