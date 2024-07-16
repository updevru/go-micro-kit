package discovery

type Discovery interface {
	RegisterService() error
	DeregisterService() error
	RegisterWorker() error
	DeregisterWorker() error
}
