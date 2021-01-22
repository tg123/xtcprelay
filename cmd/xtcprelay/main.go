package main

import (
	"log"
	"net"
	"os"

	"github.com/tg123/xtcprelay"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Usage = "make everything on earth a tcp relay"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:     "driver, d",
			Value:    "azqueue",
			Required: true,
		},
	}

	app.Flags = append(app.Flags, allFlags()...)

	app.Commands = []cli.Command{
		{
			Name:  "serverside",
			Usage: "run in server side mode, accept connections from relayer and dial to server address",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "azqueue-relayer-address",
					Required: true,
				},
				cli.StringFlag{
					Name:     "azqueue-server-address",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				d, err := getDriver(c)
				if err != nil {
					return err
				}

				r, err := d.createRelayerServerSide(c)
				if err != nil {
					return err
				}

				targetAddr := c.String("azqueue-server-address")
				relayerAddr := c.String("azqueue-relayer-address")
				log.Printf("registering to relayer with address [%v] and forward request accepted to server adderss [%v]", relayerAddr, targetAddr)

				return xtcprelay.RunRelayerServerSide(r, c.String("azqueue-relayer-address"), func(address string) (net.Conn, error) {
					return net.Dial("tcp", targetAddr)
				})
			},
		},
		{
			Name:  "clientside",
			Usage: "run in client side mode, create a local port for client app to connect and forward connections to replayer",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "azqueue-relayer-address",
					Required: true,
				},
				cli.StringFlag{
					Name:     "azqueue-listen-address",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				d, err := getDriver(c)
				if err != nil {
					return err
				}

				r, err := d.createRelayerClientSide(c)
				if err != nil {
					return err
				}

				l, err := net.Listen("tcp", c.String("azqueue-listen-address"))
				if err != nil {
					return err
				}

				relayerAddr := c.String("azqueue-relayer-address")

				log.Printf("listening on [%v] and forward request to relayer adderss [%v]", l.Addr().String(), relayerAddr)
				return xtcprelay.RunRelayerClientSide(r, l, func(address string) string {
					return relayerAddr
				})
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
