package client

import (
	"github.com/apoydence/loggrebutterfly/client/internal/filesystem"
	"github.com/apoydence/loggrebutterfly/client/internal/hasher"
	v2 "github.com/apoydence/loggrebutterfly/pb/loggregator/v2"
	"github.com/apoydence/petasos/router"
	"github.com/golang/protobuf/proto"
)

type Client struct {
	router *router.Router
}

func New(masterAddr string) *Client {
	fs := filesystem.New(masterAddr)
	hasher := hasher.New()

	router := router.New(fs, hasher)

	return &Client{
		router: router,
	}
}

func (c *Client) Write(e *v2.Envelope) error {
	data, err := proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.router.Write(data)
}
