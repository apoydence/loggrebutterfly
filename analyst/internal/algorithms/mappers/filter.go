package mappers

import (
	loggregator "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	v1 "github.com/apoydence/loggrebutterfly/api/v1"
)

type filter struct {
	info *v1.QueryInfo
}

func NewFilter(info *v1.AggregateInfo) filter {
	return filter{
		info: info.GetQuery(),
	}
}

func (f filter) Filter(e *loggregator.Envelope) (keep bool) {
	return f.info.GetFilter().GetSourceId() == e.GetSourceId() &&
		filterViaTimestamp(f.info, e) &&
		filterViaCounter(f.info, e)
}

func filterViaTimestamp(info *v1.QueryInfo, e *loggregator.Envelope) bool {
	if info.GetFilter().TimeRange == nil {
		return true
	}

	return e.Timestamp >= info.GetFilter().GetTimeRange().GetStart() &&
		e.Timestamp < info.GetFilter().GetTimeRange().GetEnd()
}

func filterViaCounter(info *v1.QueryInfo, e *loggregator.Envelope) bool {
	if info.GetFilter().GetCounter() == nil {
		return true
	}

	if e.GetCounter() == nil {
		return false
	}

	filterName := info.GetFilter().GetCounter().GetName()
	if filterName == "" {
		return true
	}

	return filterName == e.GetCounter().GetName()
}
