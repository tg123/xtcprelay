package main

import (
	"fmt"
	"net/url"

	"github.com/Azure/azure-storage-queue-go/azqueue"
	"github.com/tg123/xtcprelay"
	"github.com/tg123/xtcprelay/azure"
	"github.com/urfave/cli"
)

type azurequeue struct {
}

func (a *azurequeue) createServiceURL(c *cli.Context) (*azqueue.ServiceURL, error) {
	accountName := c.GlobalString("azqueue-account")
	accountKey := c.GlobalString("azqueue-key")

	credential, err := azqueue.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, err
	}

	p := azqueue.NewPipeline(credential, azqueue.PipelineOptions{})
	u, err := url.Parse(fmt.Sprintf("https://%s.queue.core.windows.net", accountName))
	if err != nil {
		return nil, err
	}

	serviceURL := azqueue.NewServiceURL(*u, p)
	return &serviceURL, nil
}

func (a *azurequeue) createRelayerClientSide(c *cli.Context) (xtcprelay.RelayerClientSide, error) {
	serviceURL, err := a.createServiceURL(c)
	if err != nil {
		return nil, err
	}

	return azure.NewQueueRelayerClientSide(serviceURL)
}

func (a *azurequeue) createRelayerServerSide(c *cli.Context) (xtcprelay.RelayerServerSide, error) {
	serviceURL, err := a.createServiceURL(c)
	if err != nil {
		return nil, err
	}

	return azure.NewQueueRelayerServerSide(serviceURL)
}

func (a *azurequeue) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name: "azqueue-account",
		},
		cli.StringFlag{
			Name: "azqueue-key",
		},
	}
}

func init() {
	registerDriver("azqueue", &azurequeue{})
}
