package end2end

import (
	"github.com/poy/loggrebutterfly/api/intra"
	pb "github.com/poy/talaria/api/v1"
)

//go:generate hel

type SchedulerServer interface {
	pb.SchedulerServer
}

type DataNodeServer interface {
	intra.DataNodeServer
}
