package end2end

import "github.com/apoydence/talaria/pb"

//go:generate hel

type NodeServer interface {
	pb.NodeServer
}
