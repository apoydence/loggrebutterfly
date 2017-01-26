package end2end

import (
	"github.com/apoydence/loggrebutterfly/internal/pb/intra"
	"github.com/apoydence/talaria/pb"
)

//go:generate hel

type SchedulerServer interface {
	pb.SchedulerServer
}

type RouterServer interface {
	intra.RouterServer
}
