package end2end

import (
	"github.com/apoydence/loggrebutterfly/pb/intra"
	"github.com/apoydence/talaria/pb"
)

//go:generate hel

type SchedulerServer interface {
	pb.SchedulerServer
}

type DataNodeServer interface {
	intra.DataNodeServer
}
