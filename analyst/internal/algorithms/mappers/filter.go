package mappers

import (
	"regexp"

	loggregator "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	v1 "github.com/apoydence/loggrebutterfly/api/v1"
)

type filter struct {
	info  *v1.QueryInfo
	regex *regexp.Regexp
}

func NewFilter(info *v1.AggregateInfo) (Filter, error) {
	var r *regexp.Regexp
	if pattern := info.GetQuery().GetFilter().GetLog().GetRegexp(); pattern != "" {
		var err error
		r, err = regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
	}

	return filter{
		info:  info.GetQuery(),
		regex: r,
	}, nil
}

func (f filter) Filter(e *loggregator.Envelope) (keep bool) {
	return f.info.GetFilter().GetSourceId() == e.GetSourceId() &&
		f.filterViaTimestamp(f.info, e) &&
		f.filterViaCounter(f.info, e) &&
		f.filterViaLog(f.info, e)
}

func (f filter) filterViaTimestamp(info *v1.QueryInfo, e *loggregator.Envelope) bool {
	if info.GetFilter().TimeRange == nil {
		return true
	}

	return e.Timestamp >= info.GetFilter().GetTimeRange().GetStart() &&
		e.Timestamp < info.GetFilter().GetTimeRange().GetEnd()
}

func (f filter) filterViaLog(info *v1.QueryInfo, e *loggregator.Envelope) bool {
	if info.GetFilter().GetLog() == nil {
		return true
	}

	if e.GetLog() == nil {
		return false
	}

	envPayload := e.GetLog().GetPayload()

	switch info.GetFilter().GetLog().GetPayload().(type) {
	case *v1.LogFilter_Match:
		payload := info.GetFilter().GetLog().GetMatch()
		if payload == nil {
			return true
		}

		if len(payload) != len(envPayload) {
			return false
		}

		for i := range payload {
			if payload[i] != envPayload[i] {
				return false
			}
		}

		return true
	case *v1.LogFilter_Regexp:
		pattern := info.GetFilter().GetLog().GetRegexp()
		if pattern == "" {
			return true
		}

		return f.regex.Match(envPayload)
	default:
		return true
	}

	return true
}

func (f filter) filterViaCounter(info *v1.QueryInfo, e *loggregator.Envelope) bool {
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