package filesystem

import "github.com/poy/loggrebutterfly/api/v1"

//go:generate hel

type DataNodeServer interface {
	loggrebutterfly.DataNodeServer
}

type MasterServer interface {
	loggrebutterfly.MasterServer
}
