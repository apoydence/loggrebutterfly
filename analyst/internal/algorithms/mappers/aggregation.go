package mappers

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"time"

	loggregator "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	v1 "github.com/apoydence/loggrebutterfly/api/v1"
)

type Aggregation struct {
	info   *v1.AggregateInfo
	filter Filter
}

func NewAggregation(info *v1.AggregateInfo, filter Filter) (Aggregation, error) {
	if info.GetQuery().GetFilter().GetLog() != nil {
		return Aggregation{}, fmt.Errorf("invalid filter: log")
	}

	if g := info.GetQuery().GetFilter().GetGauge(); g != nil && g.GetName() == "" {
		return Aggregation{}, fmt.Errorf("missing name field")
	}

	return Aggregation{
		info:   info,
		filter: filter,
	}, nil
}

func (a Aggregation) Map(value []byte) (key string, output []byte, err error) {
	e, err := marshalAndFilter(value, a.filter)
	if err != nil || e == nil {
		return "", nil, err
	}

	f := a.extractValue(e)
	t := time.Unix(0, e.Timestamp).
		Truncate(time.Duration(a.info.BucketWidthNs)).
		UnixNano()

	bits := math.Float64bits(f)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)

	return strconv.FormatInt(t, 10), bytes, nil
}

func (a Aggregation) extractValue(e *loggregator.Envelope) float64 {
	switch x := a.info.GetQuery().GetFilter().Envelopes.(type) {
	case *v1.AnalystFilter_Counter:
		return float64(e.GetCounter().GetTotal())
	case *v1.AnalystFilter_Gauge:
		return e.GetGauge().GetMetrics()[x.Gauge.GetName()].GetValue()
	default:
		return 0
	}
}
