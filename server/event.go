package server

type Event struct {
	ServiceStart func() error
	ServiceStop  func() error
	WorkerStart  func() error
	WorkerStop   func() error
}
