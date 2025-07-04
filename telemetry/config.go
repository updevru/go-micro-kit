package telemetry

type Config struct {
	Protocol string `env:"EXPORTER_OTLP_PROTOCOL, default=grpc"`
}

func (c *Config) IsProtocolGrpc() bool {
	return c.Protocol == "grpc"
}

func (c *Config) IsProtocolHttp() bool {
	return c.Protocol == "http/json"
}
