package main

import (
	"fmt"

	"github.com/tg123/xtcprelay"
	"github.com/urfave/cli"
)

type driver interface {
	createRelayerClientSide(c *cli.Context) (xtcprelay.RelayerClientSide, error)
	createRelayerServerSide(c *cli.Context) (xtcprelay.RelayerServerSide, error)
	flags() []cli.Flag
}

var registry map[string]driver = make(map[string]driver)

func getDriver(c *cli.Context) (driver, error) {
	d, ok := registry[c.GlobalString("driver")]
	if !ok {
		return nil, fmt.Errorf("driver not found")
	}
	return d, nil
}

func registerDriver(name string, d driver) {
	registry[name] = d
}

func allFlags() []cli.Flag {
	var flags []cli.Flag
	for _, d := range registry {
		flags = append(flags, d.flags()...)
	}

	return flags
}
