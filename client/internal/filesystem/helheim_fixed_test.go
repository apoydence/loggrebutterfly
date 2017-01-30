// This file was generated by github.com/nelsam/hel.  Do not
// edit this code by hand unless you *really* know what you're
// doing.  Expect any changes made manually to be overwritten
// the next time hel regenerates this file.

package filesystem_test

import (
	"github.com/apoydence/loggrebutterfly/pb"
	"golang.org/x/net/context"
)

type mockDataNodeServer struct {
	WriteCalled chan bool
	WriteInput  struct {
		Arg0 chan context.Context
		Arg1 chan *pb.WriteInfo
	}
	WriteOutput struct {
		Ret0 chan *pb.WriteResponse
		Ret1 chan error
	}
	ReadCalled chan bool
	ReadInput  struct {
		Arg0 chan *pb.ReadInfo
		Arg1 chan pb.DataNode_ReadServer
	}
	ReadOutput struct {
		Ret0 chan error
	}
}

func newMockDataNodeServer() *mockDataNodeServer {
	m := &mockDataNodeServer{}
	m.WriteCalled = make(chan bool, 100)
	m.WriteInput.Arg0 = make(chan context.Context, 100)
	m.WriteInput.Arg1 = make(chan *pb.WriteInfo, 100)
	m.WriteOutput.Ret0 = make(chan *pb.WriteResponse, 100)
	m.WriteOutput.Ret1 = make(chan error, 100)
	m.ReadCalled = make(chan bool, 100)
	m.ReadInput.Arg0 = make(chan *pb.ReadInfo, 100)
	m.ReadInput.Arg1 = make(chan pb.DataNode_ReadServer, 100)
	m.ReadOutput.Ret0 = make(chan error, 100)
	return m
}
func (m *mockDataNodeServer) Write(arg0 context.Context, arg1 *pb.WriteInfo) (*pb.WriteResponse, error) {
	m.WriteCalled <- true
	m.WriteInput.Arg0 <- arg0
	m.WriteInput.Arg1 <- arg1
	return <-m.WriteOutput.Ret0, <-m.WriteOutput.Ret1
}
func (m *mockDataNodeServer) Read(arg0 *pb.ReadInfo, arg1 pb.DataNode_ReadServer) error {
	m.ReadCalled <- true
	m.ReadInput.Arg0 <- arg0
	m.ReadInput.Arg1 <- arg1
	return <-m.ReadOutput.Ret0
}

type mockMasterServer struct {
	RoutesCalled chan bool
	RoutesInput  struct {
		Arg0 chan context.Context
		Arg1 chan *pb.RoutesInfo
	}
	RoutesOutput struct {
		Ret0 chan *pb.RoutesResponse
		Ret1 chan error
	}
}

func newMockMasterServer() *mockMasterServer {
	m := &mockMasterServer{}
	m.RoutesCalled = make(chan bool, 100)
	m.RoutesInput.Arg0 = make(chan context.Context, 100)
	m.RoutesInput.Arg1 = make(chan *pb.RoutesInfo, 100)
	m.RoutesOutput.Ret0 = make(chan *pb.RoutesResponse, 100)
	m.RoutesOutput.Ret1 = make(chan error, 100)
	return m
}
func (m *mockMasterServer) Routes(arg0 context.Context, arg1 *pb.RoutesInfo) (*pb.RoutesResponse, error) {
	m.RoutesCalled <- true
	m.RoutesInput.Arg0 <- arg0
	m.RoutesInput.Arg1 <- arg1
	return <-m.RoutesOutput.Ret0, <-m.RoutesOutput.Ret1
}