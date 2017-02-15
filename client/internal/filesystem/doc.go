package filesystem

import "github.com/apoydence/loggrebutterfly/api/v1"

//go:generate hel

type DataNodeServer interface {
	loggrebutterfly.DataNodeServer
}

type MasterServer interface {
	loggrebutterfly.MasterServer
}
