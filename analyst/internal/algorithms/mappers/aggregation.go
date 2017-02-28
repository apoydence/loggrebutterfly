package mappers

import (
	"encoding/binary"
	"math"
	"strconv"
	"time"

	v1 "github.com/apoydence/loggrebutterfly/api/v1"
)

type Aggregation struct {
	info   *v1.AggregateInfo
	filter Filter
}

func NewAggregation(info *v1.AggregateInfo, filter Filter) Aggregation {
	return Aggregation{
		info:   info,
		filter: filter,
	}
}

func (a Aggregation) Map(value []byte) (key string, output []byte, err error) {
	e, err := marshalAndFilter(value, a.filter)
	if err != nil || e == nil {
		return "", nil, err
	}

	c := e.GetCounter()
	t := time.Unix(0, e.Timestamp).
		Truncate(time.Duration(a.info.BucketWidthNs)).
		UnixNano()

	bits := math.Float64bits(float64(c.GetTotal()))
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)

	return strconv.FormatInt(t, 10), bytes, nil
}
