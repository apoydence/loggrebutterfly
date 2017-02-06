package networkreader

import "github.com/apoydence/loggrebutterfly/api/intra"

//go:generate hel

type DataNodeServer interface {
	intra.DataNodeServer
}
