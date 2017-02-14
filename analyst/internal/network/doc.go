package network

import (
	"github.com/apoydence/loggrebutterfly/api/intra"
)

//go:generate hel

type AnalystServer interface {
	intra.AnalystServer
}
