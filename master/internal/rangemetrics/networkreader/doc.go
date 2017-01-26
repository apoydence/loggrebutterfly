package networkreader

import "github.com/apoydence/loggrebutterfly/pb/intra"

//go:generate hel

type DataNodeServer interface {
	intra.DataNodeServer
}
