package networkreader

import "github.com/poy/loggrebutterfly/api/intra"

//go:generate hel

type DataNodeServer interface {
	intra.DataNodeServer
}
