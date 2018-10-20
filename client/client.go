package client

import (
	"io"

	v2 "github.com/poy/loggrebutterfly/api/loggregator/v2"
	"github.com/poy/loggrebutterfly/client/internal/filesystem"
	"github.com/poy/loggrebutterfly/client/internal/hasher"
	"github.com/poy/petasos/reader"
	"github.com/poy/petasos/router"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

type Client struct {
	router *router.Router
	reader *reader.RouteReader
	hasher *hasher.Hasher
}

func New(masterAddr string) *Client {
	cache := filesystem.NewCache(masterAddr)
	fs := filesystem.New(cache)
	hasher := hasher.New()

	counter := router.NewCounter()
	router := router.New(fs, hasher, counter)

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

type DataPacket struct {
	Envelope *v2.Envelope
	Filename string
	Index    uint64
}

func (c *Client) ReadFrom(sourceID string) (func() (DataPacket, error), error) {
	hash := c.hasher.HashString(sourceID)

	r := c.reader.ReadFrom(hash)
	return func() (DataPacket, error) {
		for {
			data, err := r.Read()
			if grpc.ErrorDesc(err) == "EOF" {
				return DataPacket{}, io.EOF
			}

			if err != nil {
				return DataPacket{}, err
			}

			var e v2.Envelope
			if err := proto.Unmarshal(data.Payload, &e); err != nil {
				return DataPacket{}, err
			}

			if e.SourceId != sourceID {
				continue
			}

			return DataPacket{
				Envelope: &e,
				Filename: data.Filename,
				Index:    data.Index,
			}, nil
		}
	}, nil
}
