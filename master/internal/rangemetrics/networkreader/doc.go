package networkreader

import "github.com/apoydence/loggrebutterfly/internal/pb/intra"

//go:generate hel
type RouterServer interface {
	intra.RouterServer
}
