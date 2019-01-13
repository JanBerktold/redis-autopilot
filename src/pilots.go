package main

type Pilot interface {
	Execute() error
	Shutdown()
}

type masterWithSlavesPilot struct {
	redisInstanceProvider  RedisInstanceProvider
	consulServiceRegistrar ConsulServiceRegistrar
}

func NewMasterWithSlavesPilot(redisInstanceProvider RedisInstanceProvider, consulServiceRegistrar ConsulServiceRegistrar) Pilot {
	return &masterWithSlavesPilot{
		redisInstanceProvider:  redisInstanceProvider,
		consulServiceRegistrar: consulServiceRegistrar,
	}
}

func (pilot *masterWithSlavesPilot) Execute() error {
	instance := pilot.redisInstanceProvider.Instance()

	redisStatus := instance.Status()
	if redisStatus != RedisStatusReady {
		// make sure we're not in consul
	}

	return nil
}

func (p *masterWithSlavesPilot) Shutdown() {

}
