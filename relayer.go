package xtcprelay

import (
	"io"
	"log"
	"net"
)

type Relayer interface {
	RelayerServerSide
	RelayerClientSide
}

type RelayerServerSide interface {
	io.Closer
	Listen(address string) (net.Listener, error)
}

type RelayerClientSide interface {
	io.Closer
	Dial(address string) (net.Conn, error)
}

func copy(dst, src io.ReadWriteCloser) error {

	c := make(chan error, 2)

	go func() {
		_, err := io.Copy(dst, src)
		c <- err
	}()

	go func() {
		_, err := io.Copy(src, dst)
		c <- err
	}()

	defer src.Close()
	defer dst.Close()

	return <-c
}

func relay(l net.Listener, dialer func(address string) (net.Conn, error)) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}

		transit, err := dialer(conn.RemoteAddr().String())
		if err != nil {
			log.Printf("failed to dial relay connection: %v", err)
			continue
		}

		go copy(conn, transit)
	}
}

func RunRelayerClientSide(r RelayerClientSide, l net.Listener, addressmapper func(address string) string) error {
	return relay(l, func(address string) (net.Conn, error) {
		return r.Dial(addressmapper(address))
	})
}

func RunRelayerServerSide(r RelayerServerSide, listenAddress string, dialer func(address string) (net.Conn, error)) error {
	l, err := r.Listen(listenAddress)

	if err != nil {
		return err
	}

	return relay(l, dialer)
}
