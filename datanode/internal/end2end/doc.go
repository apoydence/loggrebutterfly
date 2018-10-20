package end2end

import pb "github.com/poy/talaria/api/v1"

//go:generate hel

type NodeServer interface {
	pb.NodeServer
}
