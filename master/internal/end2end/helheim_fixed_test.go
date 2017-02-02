// This file was generated by github.com/nelsam/hel.  Do not
// edit this code by hand unless you *really* know what you're
// doing.  Expect any changes made manually to be overwritten
// the next time hel regenerates this file.

package end2end_test

import (
	"github.com/apoydence/loggrebutterfly/pb/intra"
	pb "github.com/apoydence/talaria/api/v1"
	"golang.org/x/net/context"
)

type mockSchedulerServer struct {
	CreateCalled chan bool
	CreateInput  struct {
		Arg0 chan context.Context
		Arg1 chan *pb.CreateInfo
	}
	CreateOutput struct {
		Ret0 chan *pb.CreateResponse
		Ret1 chan error
	}
	ReadOnlyCalled chan bool
	ReadOnlyInput  struct {
		Arg0 chan context.Context
		Arg1 chan *pb.ReadOnlyInfo
	}
	ReadOnlyOutput struct {
		Ret0 chan *pb.ReadOnlyResponse
		Ret1 chan error
	}
	ListClusterInfoCalled chan bool
	ListClusterInfoInput  struct {
		Arg0 chan context.Context
		Arg1 chan *pb.ListInfo
	}
	ListClusterInfoOutput struct {
		Ret0 chan *pb.ListResponse
		Ret1 chan error
	}
}

func newMockSchedulerServer() *mockSchedulerServer {
	m := &mockSchedulerServer{}
	m.CreateCalled = make(chan bool, 100)
	m.CreateInput.Arg0 = make(chan context.Context, 100)
	m.CreateInput.Arg1 = make(chan *pb.CreateInfo, 100)
	m.CreateOutput.Ret0 = make(chan *pb.CreateResponse, 100)
	m.CreateOutput.Ret1 = make(chan error, 100)
	m.ReadOnlyCalled = make(chan bool, 100)
	m.ReadOnlyInput.Arg0 = make(chan context.Context, 100)
	m.ReadOnlyInput.Arg1 = make(chan *pb.ReadOnlyInfo, 100)
	m.ReadOnlyOutput.Ret0 = make(chan *pb.ReadOnlyResponse, 100)
	m.ReadOnlyOutput.Ret1 = make(chan error, 100)
	m.ListClusterInfoCalled = make(chan bool, 100)
	m.ListClusterInfoInput.Arg0 = make(chan context.Context, 100)
	m.ListClusterInfoInput.Arg1 = make(chan *pb.ListInfo, 100)
	m.ListClusterInfoOutput.Ret0 = make(chan *pb.ListResponse, 100)
	m.ListClusterInfoOutput.Ret1 = make(chan error, 100)
	return m
}
func (m *mockSchedulerServer) Create(arg0 context.Context, arg1 *pb.CreateInfo) (*pb.CreateResponse, error) {
	m.CreateCalled <- true
	m.CreateInput.Arg0 <- arg0
	m.CreateInput.Arg1 <- arg1
	return <-m.CreateOutput.Ret0, <-m.CreateOutput.Ret1
}
func (m *mockSchedulerServer) ReadOnly(arg0 context.Context, arg1 *pb.ReadOnlyInfo) (*pb.ReadOnlyResponse, error) {
	m.ReadOnlyCalled <- true
	m.ReadOnlyInput.Arg0 <- arg0
	m.ReadOnlyInput.Arg1 <- arg1
	return <-m.ReadOnlyOutput.Ret0, <-m.ReadOnlyOutput.Ret1
}
func (m *mockSchedulerServer) ListClusterInfo(arg0 context.Context, arg1 *pb.ListInfo) (*pb.ListResponse, error) {
	m.ListClusterInfoCalled <- true
	m.ListClusterInfoInput.Arg0 <- arg0
	m.ListClusterInfoInput.Arg1 <- arg1
	return <-m.ListClusterInfoOutput.Ret0, <-m.ListClusterInfoOutput.Ret1
}

type mockDataNodeServer struct {
	ReadMetricsCalled chan bool
	ReadMetricsInput  struct {
		Arg0 chan context.Context
		Arg1 chan *intra.ReadMetricsInfo
	}
	ReadMetricsOutput struct {
		Ret0 chan *intra.ReadMetricsResponse
		Ret1 chan error
	}
}

func newMockDataNodeServer() *mockDataNodeServer {
	m := &mockDataNodeServer{}
	m.ReadMetricsCalled = make(chan bool, 100)
	m.ReadMetricsInput.Arg0 = make(chan context.Context, 100)
	m.ReadMetricsInput.Arg1 = make(chan *intra.ReadMetricsInfo, 100)
	m.ReadMetricsOutput.Ret0 = make(chan *intra.ReadMetricsResponse, 100)
	m.ReadMetricsOutput.Ret1 = make(chan error, 100)
	return m
}
func (m *mockDataNodeServer) ReadMetrics(arg0 context.Context, arg1 *intra.ReadMetricsInfo) (*intra.ReadMetricsResponse, error) {
	m.ReadMetricsCalled <- true
	m.ReadMetricsInput.Arg0 <- arg0
	m.ReadMetricsInput.Arg1 <- arg1
	return <-m.ReadMetricsOutput.Ret0, <-m.ReadMetricsOutput.Ret1
}
