package filesystem

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "github.com/apoydence/loggrebutterfly/pb/v1"
)

type Cache struct {
	masterClient pb.MasterClient

	mu     sync.Mutex
	routes map[string]clientInfo
}

func NewCache(masterAddr string) *Cache {
	return &Cache{
		masterClient: setupMasterClient(masterAddr),
	}
}

func (c *Cache) FetchRoute(name string) (client pb.DataNodeClient, addr string) {
	info, ok := c.fetchRoute(name)
	if !ok {
		return nil, ""
	}

	return info.client, info.addr
}

func (c *Cache) Reset() {
	log.Printf("Resting route cache...")
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, r := range c.routes {
		r.closer.Close()
	}

	c.routes = nil
}

func (c *Cache) List() (files []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.setupRoutes(); err != nil {
		log.Printf("Failed to setup routes: %s", err)
		return nil
	}

	for name, _ := range c.routes {
		files = append(files, name)
	}

	return files
}

type clientInfo struct {
	addr   string
	client pb.DataNodeClient
	closer io.Closer
}

func (c *Cache) fetchRoute(name string) (clientInfo, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.setupRoutes(); err != nil {
		log.Printf("Failed to setup routes: %s", err)
		return clientInfo{}, false
	}

	client, ok := c.routes[name]
	return client, ok
}

func (c *Cache) setupRoutes() error {
	if c.routes != nil {
		return nil
	}

	files, err := c.list()
	if err != nil {
		return err
	}

	c.routes = make(map[string]clientInfo)

	for _, file := range files {
		c.routes[file.Name] = setupDataClient(file.Leader)
	}

	return nil
}

func (c *Cache) list() (files []*pb.RouteInfo, err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := c.masterClient.Routes(ctx, new(pb.RoutesInfo))
	if err != nil {
		return nil, err
	}

	return resp.Routes, nil
}

func setupMasterClient(addr string) pb.MasterClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to master: %s", err)
	}
	return pb.NewMasterClient(conn)
}

func setupDataClient(addr string) clientInfo {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect to master: %s", err)
	}
	return clientInfo{
		addr:   addr,
		client: pb.NewDataNodeClient(conn),
		closer: conn,
	}
}
