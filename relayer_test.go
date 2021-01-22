package xtcprelay

import (
	"net"
	"testing"
	"time"
)

type tcpRelayer struct {
	server net.Listener
}

func (r *tcpRelayer) Close() error {
	return nil
}

func (r *tcpRelayer) Listen(address string) (net.Listener, error) {
	return r.server, nil
}

func (r *tcpRelayer) Dial(address string) (net.Conn, error) {
	return net.Dial("tcp", r.server.Addr().String())
}

func newTcpRelayer() *tcpRelayer {
	l, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		panic(err)
	}

	return &tcpRelayer{l}
}

func TestRelay(t *testing.T) {

	realserver, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("cannot create server %v", err)
	}

	go func() {

		for {
			c, err := realserver.Accept()

			if err != nil {
				t.Logf("accept failed %v", err)
				continue
			}

			go func(conn net.Conn) {
				defer conn.Close()

				data := make([]byte, 1)

				for {
					_, err := conn.Read(data)

					if err != nil {
						t.Logf("read err%v", err)
						return
					}

					data[0]++

					_, err = conn.Write(data)

					if err != nil {
						t.Logf("read err%v", err)
						return
					}
				}

			}(c)
		}
	}()

	tests := []struct {
		name  string
		relay Relayer
	}{
		{
			name:  "tcprelay",
			relay: newTcpRelayer(),
		},
		{
			name:  "memrelay",
			relay: newMemRelayer(),
		},
	}

	for _, tt := range tests {
		// serve a relay to real server
		go RunRelayerServerSide(tt.relay, "", func(address string) (net.Conn, error) {
			return net.Dial("tcp", realserver.Addr().String())
		})

		// create a dummy server to relay
		s, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("create tcp relay %v", err)
		}
		go RunRelayerClientSide(tt.relay, s, func(address string) string { return address })

		c, err := net.Dial("tcp", s.Addr().String())
		if err != nil {
			t.Fatalf("connect dummy svr %v", err)
		}

		for i := 1; i < 100; i++ {
			c.Write([]byte{byte(i)})

			b := make([]byte, 1)
			c.Read(b)

			// fmt.Println(b[0], i)
			if int(b[0]) != i+1 {
				t.Errorf("%v data failure ", tt.name)
			}
		}
	}

}

type memRelayer struct {
	l *memListener
}

func (r *memRelayer) Close() error {
	return r.l.Close()
}

type memRelayAddr struct {
}

func (a *memRelayAddr) Network() string {
	return "xrelay-mem"
}

func (a *memRelayAddr) String() string {
	return ""
}

type memConn struct {
	input  chan byte
	output chan byte
}

func (c *memConn) Read(b []byte) (n int, err error) {
	b[0] = <-c.output
	return 1, nil
}

func (c *memConn) Write(b []byte) (n int, err error) {
	for i := 0; i < len(b); i++ {
		c.input <- b[i]
	}
	return len(b), nil
}

func (c *memConn) Close() error {
	return nil
}

func (c *memConn) LocalAddr() net.Addr {
	return &memRelayAddr{}
}

func (c *memConn) RemoteAddr() net.Addr {
	return &memRelayAddr{}
}

func (c *memConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *memConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *memConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type memListener struct {
	conns chan *memConn
}

func (l *memListener) Accept() (net.Conn, error) {
	c := <-l.conns
	return &memConn{
		input:  c.output,
		output: c.input,
	}, nil
}

func (l *memListener) Close() error {
	return nil
}

func (l *memListener) Addr() net.Addr {
	return &memRelayAddr{}
}

func (r *memRelayer) Listen(address string) (net.Listener, error) {
	return r.l, nil
}

func (r *memRelayer) Dial(address string) (net.Conn, error) {
	c := &memConn{
		input:  make(chan byte),
		output: make(chan byte),
	}

	r.l.conns <- c
	return c, nil
}

func newMemRelayer() *memRelayer {
	return &memRelayer{
		l: &memListener{
			conns: make(chan *memConn),
		},
	}
}
