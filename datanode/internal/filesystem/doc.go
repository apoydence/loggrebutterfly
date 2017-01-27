package filesystem

import "github.com/apoydence/talaria/pb"

//go:generate hel
//
type NodeServer interface {
	pb.NodeServer
}
