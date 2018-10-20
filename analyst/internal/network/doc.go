package network

import (
	"github.com/poy/loggrebutterfly/api/intra"
)

//go:generate hel

type AnalystServer interface {
	intra.AnalystServer
}
