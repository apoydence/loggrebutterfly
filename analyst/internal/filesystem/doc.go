package filesystem

import (
	talaria "github.com/poy/talaria/api/v1"
)

//go:generate hel

type SchedulerClient interface {
	talaria.SchedulerClient
}

type NodeClient interface {
	talaria.NodeClient
}

type NodeReadClient interface {
	talaria.Node_ReadClient
}
