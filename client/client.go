package client

import (
	"io"

	v2 "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	"github.com/apoydence/loggrebutterfly/client/internal/filesystem"
	"github.com/apoydence/loggrebutterfly/client/internal/hasher"
	"github.com/apoydence/petasos/reader"
	"github.com/apoydence/petasos/router"
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

type ReadFromOption func(c *readConf)

type DataPacket struct {
	Envelope *v2.Envelope
	Filename string
	Index    uint64
}

func (c *Client) ReadFrom(sourceID string, opts ...ReadFromOption) (func() (DataPacket, error), error) {
	var conf readConf
	for _, o := range opts {
		o(&conf)
	}

	hash := c.hasher.HashString(sourceID)

	r, err := c.fetchReader(conf.startFileName, hash, conf.startIndex)
	if err != nil {
		return nil, err
	}

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

type readConf struct {
	startFileName string
	startIndex    uint64
}

func WithStartingFileName(fileName string) ReadFromOption {
	return func(c *readConf) {
		c.startFileName = fileName
	}
}

func WithStartingIndex(startIndex uint64) ReadFromOption {
	return func(c *readConf) {
		c.startIndex = startIndex
	}
}

func (c *Client) fetchReader(fileName string, hash, startingIndex uint64) (reader.Reader, error) {
	if fileName == "" {
		return c.reader.ReadFrom(hash), nil
	}

	return c.reader.ReadFromMidStream(hash, fileName, startingIndex)
}
