package filesystem

import pb "github.com/apoydence/loggrebutterfly/api/v1"

//go:generate hel

type DataNodeServer interface {
	pb.DataNodeServer
}

type MasterServer interface {
	pb.MasterServer
}
