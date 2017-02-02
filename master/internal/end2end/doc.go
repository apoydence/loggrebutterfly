package end2end

import (
	"github.com/apoydence/loggrebutterfly/pb/intra"
	pb "github.com/apoydence/talaria/api/v1"
)

//go:generate hel

type SchedulerServer interface {
	pb.SchedulerServer
}

type DataNodeServer interface {
	intra.DataNodeServer
}
