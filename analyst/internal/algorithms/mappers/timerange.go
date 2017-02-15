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

func NewTimeRange(info *v1.QueryInfo) TimeRange {
	return TimeRange{
		info: info,
	}
}

func (r TimeRange) Map(value []byte) (key string, output []byte, err error) {
	var e loggregator.Envelope
	if err := proto.Unmarshal(value, &e); err != nil {
		return "", nil, err
	}

	if r.info.SourceId != e.SourceId ||
		(e.Timestamp < r.info.TimeRange.Start || e.Timestamp >= r.info.TimeRange.End) {
		return "", nil, nil
	}

	return strconv.FormatInt(e.Timestamp, 10), value, nil
}
