package filesystem

import "github.com/apoydence/loggrebutterfly/pb"

//go:generate hel

type DataNodeServer interface {
	pb.DataNodeServer
}

type MasterServer interface {
	pb.MasterServer
}
