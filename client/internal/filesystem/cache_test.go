package filesystem_test

import (
	"net"
	"testing"

	"google.golang.org/grpc"

	"github.com/poy/eachers/testhelpers"
	"github.com/poy/loggrebutterfly/client/internal/filesystem"
	pb "github.com/poy/loggrebutterfly/api/v1"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
)

type TC struct {
	*testing.T
	c *filesystem.Cache

	mockDataNodeAddrs []string
	mockDataNodes     []*mockDataNodeServer

	mockMasterNodeAddr string
	mockMasterNode     *mockMasterServer
}

func TestCache(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TC {
		var (
			mockDataNodes     []*mockDataNodeServer
			mockDataNodeAddrs []string
		)
		for i := 0; i < 3; i++ {
			addr, m := startMockDataNode()
			mockDataNodes = append(mockDataNodes, m)
			mockDataNodeAddrs = append(mockDataNodeAddrs, addr)
		}

		mockMasterNodeAddr, mockMasterNode := startMockMaster()

		writeRoutes(mockDataNodeAddrs, mockMasterNode)

		return TC{
			T: t,
			c: filesystem.NewCache(mockMasterNodeAddr),

			mockDataNodes:      mockDataNodes,
			mockDataNodeAddrs:  mockDataNodeAddrs,
			mockMasterNode:     mockMasterNode,
			mockMasterNodeAddr: mockMasterNodeAddr,
		}
	})

	o.Spec("it reuses connections", func(t TC) {
		clientA1, addrA1 := t.c.FetchRoute("some-name-a")
		clientA2, addrA2 := t.c.FetchRoute("some-name-a")
		clientB, _ := t.c.FetchRoute("some-name-b")

		Expect(t, clientA1).To(Equal(clientA2))
		Expect(t, addrA1).To(Equal(addrA2))
		Expect(t, clientA1).To(Not(Equal(clientB)))
	})

	o.Spec("it resets the connections", func(t TC) {
		clientA, _ := t.c.FetchRoute("some-name-a")
		t.c.Reset()
		clientB, _ := t.c.FetchRoute("some-name-a")

		Expect(t, clientA).To(Not(Equal(clientB)))
	})

	o.Spec("it lists all the addrs", func(t TC) {
		list := t.c.List()

		Expect(t, list).To(HaveLen(2))
		Expect(t, list).To(Contain())
	})
}

func writeRoutes(addrs []string, master *mockMasterServer) {
	testhelpers.AlwaysReturn(master.RoutesOutput.Ret0, &pb.RoutesResponse{
		Routes: []*pb.RouteInfo{
			{
				Name:   "some-name-a",
				Leader: addrs[0],
			},
			{
				Name:   "some-name-b",
				Leader: addrs[1],
			},
		},
	})
	close(master.RoutesOutput.Ret1)
}

func startMockMaster() (string, *mockMasterServer) {
	mockMasterServer := newMockMasterServer()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterMasterServer(s, mockMasterServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return lis.Addr().String(), mockMasterServer
}
