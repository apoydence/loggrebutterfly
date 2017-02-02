package filesystem

import pb "github.com/apoydence/talaria/api/v1"

//go:generate hel
//
type NodeServer interface {
	pb.NodeServer
}
