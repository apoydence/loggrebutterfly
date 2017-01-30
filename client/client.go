package client

import (
	"github.com/apoydence/loggrebutterfly/client/internal/filesystem"
	"github.com/apoydence/loggrebutterfly/client/internal/hasher"
	v2 "github.com/apoydence/loggrebutterfly/pb/loggregator/v2"
	"github.com/apoydence/petasos/reader"
	"github.com/apoydence/petasos/router"
	"github.com/golang/protobuf/proto"
)

type Client struct {
	router *router.Router
	reader *reader.RouteReader
	hasher *hasher.Hasher
}

func New(masterAddr string) *Client {
	fs := filesystem.New(masterAddr)
	hasher := hasher.New()

	router := router.New(fs, hasher)
	reader := reader.NewRouteReader(fs)

	return &Client{
		hasher: hasher,
		router: router,
		reader: reader,
	}
}

func (c *Client) Write(e *v2.Envelope) error {
	data, err := proto.Marshal(e)
	if err != nil {
		return err
	}

	return c.router.Write(data)
}

func (c *Client) ReadFrom(sourceUUID string) func() (*v2.Envelope, error) {
	hash := c.hasher.HashString(sourceUUID)

	r := c.reader.ReadFrom(hash)
	return func() (*v2.Envelope, error) {
		data, err := r.Read()
		if err != nil {
			return nil, err
		}

		var e v2.Envelope
		if err := proto.Unmarshal(data, &e); err != nil {
			return nil, err
		}

		return &e, nil
	}
}
