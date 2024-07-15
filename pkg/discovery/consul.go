package discovery

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/updevru/go-micro-kit/pkg/config"
	"strconv"
)

type Consul struct {
	serviceName string
	workerName  string
	configHttp  *config.Http
	configGrpc  *config.Grpc
	client      *api.Client
}

func NewConsul(config *config.App, configHttp *config.Http, configGrpc *config.Grpc) (*Consul, error) {
	consulConfig := api.DefaultConfig()
	consulClient, err := api.NewClient(consulConfig)

	return &Consul{
		serviceName: config.AppName,
		workerName:  config.AppName + "-worker",
		configHttp:  configHttp,
		configGrpc:  configGrpc,
		client:      consulClient,
	}, err
}

func (c *Consul) RegisterService() error {
	registration := &api.AgentServiceRegistration{
		ID:              c.serviceName,
		Name:            c.serviceName,
		Checks:          make(api.AgentServiceChecks, 0),
		TaggedAddresses: make(map[string]api.ServiceAddress),
		Meta:            make(map[string]string),
	}

	if c.configHttp != nil {
		port, _ := strconv.Atoi(c.configHttp.Port)
		registration.Address = c.configHttp.Host
		registration.Port = port
		registration.TaggedAddresses["http"] = api.ServiceAddress{
			Address: c.configHttp.Host,
			Port:    port,
		}
		registration.Meta["http"] = fmt.Sprintf("%s:%d", c.configHttp.Host, port)
		registration.Checks = append(registration.Checks, &api.AgentServiceCheck{
			Name:                           "http",
			HTTP:                           fmt.Sprintf("http://%s:%d/healthz", c.configHttp.Host, port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		})
	}

	if c.configGrpc != nil {
		port, _ := strconv.Atoi(c.configGrpc.Port)
		registration.Address = c.configGrpc.Host
		registration.Port = port
		registration.TaggedAddresses["grpc"] = api.ServiceAddress{
			Address: c.configGrpc.Host,
			Port:    port,
		}
		registration.Meta["grpc"] = fmt.Sprintf("%s:%d", c.configGrpc.Host, port)
		registration.Checks = append(registration.Checks, &api.AgentServiceCheck{
			Name:                           "grpc",
			GRPC:                           fmt.Sprintf("%s:%d", c.configGrpc.Host, port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		})
	}

	if err := c.client.Agent().ServiceRegister(registration); err != nil {
		return err
	}

	return nil
}

func (c *Consul) DeregisterService() error {
	return c.client.Agent().ServiceDeregister(c.serviceName)
}

func (c *Consul) RegisterWorker() error {
	registration := &api.AgentServiceRegistration{
		ID:   c.workerName,
		Name: c.workerName,
	}

	if err := c.client.Agent().ServiceRegister(registration); err != nil {
		return err
	}

	return nil
}

func (c *Consul) DeregisterWorker() error {
	return c.client.Agent().ServiceDeregister(c.workerName)
}
