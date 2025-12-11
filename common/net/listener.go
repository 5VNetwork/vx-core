package net

import (
	"net"
	"sync/atomic"
)

type StatsListener struct {
	net.Listener
	readCounter  *atomic.Uint64
	writeCounter *atomic.Uint64
}

func NewStatsListener(listener net.Listener, readCounter, writeCounter *atomic.Uint64) *StatsListener {
	return &StatsListener{
		Listener:     listener,
		readCounter:  readCounter,
		writeCounter: writeCounter,
	}
}

func (l *StatsListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewStatsConn(conn, l.readCounter, l.writeCounter), nil
}
