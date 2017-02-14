package end2end

import talaria "github.com/apoydence/talaria/api/v1"

//go:generate hel

type SchedulerServer interface {
	talaria.SchedulerServer
}

type NodeServer interface {
	talaria.NodeServer
}
