package mappers

import (
	"strconv"

	loggregator "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	v1 "github.com/apoydence/loggrebutterfly/api/v1"
	"github.com/golang/protobuf/proto"
)

type TimeRange struct {
	info *v1.QueryInfo
}

func NewTimeRange(info *v1.AggregateInfo) TimeRange {
	return TimeRange{
		info: info.GetQuery(),
	}
}

func (r TimeRange) Map(value []byte) (key string, output []byte, err error) {
	e, err := marshalAndFilter(value, r.info)
	if err != nil || e == nil {
		return "", nil, err
	}

	return strconv.FormatInt(e.Timestamp, 10), value, nil
}

func marshalAndFilter(value []byte, info *v1.QueryInfo) (*loggregator.Envelope, error) {
	var e loggregator.Envelope
	if err := proto.Unmarshal(value, &e); err != nil {
		return nil, err
	}

	if info.GetFilter().GetSourceId() != e.SourceId ||
		(e.Timestamp < info.GetFilter().GetTimeRange().Start ||
			e.Timestamp >= info.GetFilter().GetTimeRange().GetEnd()) {
		return nil, nil
	}

	return &e, nil
}
