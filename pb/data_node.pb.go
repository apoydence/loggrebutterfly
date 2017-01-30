// Code generated by protoc-gen-go.
// source: data_node.proto
// DO NOT EDIT!

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	data_node.proto
	master.proto

It has these top-level messages:
	WriteInfo
	WriteResponse
	ReadInfo
	ReadData
	RoutesInfo
	RoutesResponse
	RouteInfo
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type WriteInfo struct {
	Payload []byte `protobuf:"bytes,1,opt,name=Payload,json=payload,proto3" json:"Payload,omitempty"`
}

func (m *WriteInfo) Reset()                    { *m = WriteInfo{} }
func (m *WriteInfo) String() string            { return proto.CompactTextString(m) }
func (*WriteInfo) ProtoMessage()               {}
func (*WriteInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *WriteInfo) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

type WriteResponse struct {
}

func (m *WriteResponse) Reset()                    { *m = WriteResponse{} }
func (m *WriteResponse) String() string            { return proto.CompactTextString(m) }
func (*WriteResponse) ProtoMessage()               {}
func (*WriteResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type ReadInfo struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *ReadInfo) Reset()                    { *m = ReadInfo{} }
func (m *ReadInfo) String() string            { return proto.CompactTextString(m) }
func (*ReadInfo) ProtoMessage()               {}
func (*ReadInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ReadInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type ReadData struct {
	Payload []byte `protobuf:"bytes,1,opt,name=Payload,json=payload,proto3" json:"Payload,omitempty"`
}

func (m *ReadData) Reset()                    { *m = ReadData{} }
func (m *ReadData) String() string            { return proto.CompactTextString(m) }
func (*ReadData) ProtoMessage()               {}
func (*ReadData) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *ReadData) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func init() {
	proto.RegisterType((*WriteInfo)(nil), "pb.WriteInfo")
	proto.RegisterType((*WriteResponse)(nil), "pb.WriteResponse")
	proto.RegisterType((*ReadInfo)(nil), "pb.ReadInfo")
	proto.RegisterType((*ReadData)(nil), "pb.ReadData")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for DataNode service

type DataNodeClient interface {
	Write(ctx context.Context, in *WriteInfo, opts ...grpc.CallOption) (*WriteResponse, error)
	Read(ctx context.Context, in *ReadInfo, opts ...grpc.CallOption) (DataNode_ReadClient, error)
}

type dataNodeClient struct {
	cc *grpc.ClientConn
}

func NewDataNodeClient(cc *grpc.ClientConn) DataNodeClient {
	return &dataNodeClient{cc}
}

func (c *dataNodeClient) Write(ctx context.Context, in *WriteInfo, opts ...grpc.CallOption) (*WriteResponse, error) {
	out := new(WriteResponse)
	err := grpc.Invoke(ctx, "/pb.DataNode/Write", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dataNodeClient) Read(ctx context.Context, in *ReadInfo, opts ...grpc.CallOption) (DataNode_ReadClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_DataNode_serviceDesc.Streams[0], c.cc, "/pb.DataNode/Read", opts...)
	if err != nil {
		return nil, err
	}
	x := &dataNodeReadClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type DataNode_ReadClient interface {
	Recv() (*ReadData, error)
	grpc.ClientStream
}

type dataNodeReadClient struct {
	grpc.ClientStream
}

func (x *dataNodeReadClient) Recv() (*ReadData, error) {
	m := new(ReadData)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for DataNode service

type DataNodeServer interface {
	Write(context.Context, *WriteInfo) (*WriteResponse, error)
	Read(*ReadInfo, DataNode_ReadServer) error
}

func RegisterDataNodeServer(s *grpc.Server, srv DataNodeServer) {
	s.RegisterService(&_DataNode_serviceDesc, srv)
}

func _DataNode_Write_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WriteInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataNodeServer).Write(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.DataNode/Write",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataNodeServer).Write(ctx, req.(*WriteInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _DataNode_Read_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ReadInfo)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DataNodeServer).Read(m, &dataNodeReadServer{stream})
}

type DataNode_ReadServer interface {
	Send(*ReadData) error
	grpc.ServerStream
}

type dataNodeReadServer struct {
	grpc.ServerStream
}

func (x *dataNodeReadServer) Send(m *ReadData) error {
	return x.ServerStream.SendMsg(m)
}

var _DataNode_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.DataNode",
	HandlerType: (*DataNodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Write",
			Handler:    _DataNode_Write_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Read",
			Handler:       _DataNode_Read_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "data_node.proto",
}

func init() { proto.RegisterFile("data_node.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 184 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x4f, 0x49, 0x2c, 0x49,
	0x8c, 0xcf, 0xcb, 0x4f, 0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2a, 0x48, 0x52,
	0x52, 0xe5, 0xe2, 0x0c, 0x2f, 0xca, 0x2c, 0x49, 0xf5, 0xcc, 0x4b, 0xcb, 0x17, 0x92, 0xe0, 0x62,
	0x0f, 0x48, 0xac, 0xcc, 0xc9, 0x4f, 0x4c, 0x91, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x09, 0x62, 0x2f,
	0x80, 0x70, 0x95, 0xf8, 0xb9, 0x78, 0xc1, 0xca, 0x82, 0x52, 0x8b, 0x0b, 0xf2, 0xf3, 0x8a, 0x53,
	0x95, 0xe4, 0xb8, 0x38, 0x82, 0x52, 0x13, 0x53, 0xc0, 0xda, 0x84, 0xb8, 0x58, 0xf2, 0x12, 0x73,
	0x53, 0xc1, 0x7a, 0x38, 0x83, 0xc0, 0x6c, 0x25, 0x15, 0x88, 0xbc, 0x4b, 0x62, 0x49, 0x22, 0x6e,
	0x63, 0x8d, 0xe2, 0xb9, 0x38, 0x40, 0x2a, 0xfc, 0xf2, 0x53, 0x52, 0x85, 0xb4, 0xb9, 0x58, 0xc1,
	0x56, 0x08, 0xf1, 0xea, 0x15, 0x24, 0xe9, 0xc1, 0x1d, 0x25, 0x25, 0x08, 0xe7, 0xc2, 0x2d, 0x67,
	0x10, 0x52, 0xe3, 0x62, 0x01, 0x19, 0x2f, 0xc4, 0x03, 0x92, 0x84, 0x39, 0x44, 0x0a, 0xce, 0x03,
	0x19, 0xaa, 0xc4, 0x60, 0xc0, 0x98, 0xc4, 0x06, 0xf6, 0xa9, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff,
	0xf1, 0xe8, 0xd3, 0x5e, 0xfc, 0x00, 0x00, 0x00,
}