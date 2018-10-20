package filesystem

import pb "github.com/poy/talaria/api/v1"

//go:generate hel
//
type NodeServer interface {
	pb.NodeServer
}
