package common

type Runnable interface {
	Start() error
	Close() error
}

type Startable interface {
	Start() error
}

type Closable interface {
	Close() error
}

type Interruptible interface {
	Interrupt()
}

type HasType interface {
	// Type returns the type of the object.
	// Usually it returns (*Type)(nil) of the object.
	Type() interface{}
}
