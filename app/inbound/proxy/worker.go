package proxy

type worker interface {
	Start() error
	Close() error
}
