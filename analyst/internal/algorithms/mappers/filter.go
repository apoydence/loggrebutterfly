package mappers

import (
	"fmt"
	"regexp"

	loggregator "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	v1 "github.com/apoydence/loggrebutterfly/api/v1"
)

type filter struct {
	info  *v1.QueryInfo
	regex *regexp.Regexp
}

func NewFilter(info *v1.AggregateInfo) (Filter, error) {
	f := filter{
		info: info.GetQuery(),
	}

	if err := f.validateFilter(info); err != nil {
		return nil, err
	}

	return f, nil
}

func (f filter) Filter(e *loggregator.Envelope) (keep bool) {
	return f.info.GetFilter().GetSourceId() == e.GetSourceId() &&
		f.filterViaTimestamp(f.info, e) &&
		f.filterViaCounter(f.info, e) &&
		f.filterViaLog(f.info, e) &&
		f.filterViaGauge(f.info, e)
}

func (f *filter) validateFilter(info *v1.AggregateInfo) error {
	var r *regexp.Regexp
	if pattern := info.GetQuery().GetFilter().GetLog().GetRegexp(); pattern != "" {
		var err error
		r, err = regexp.Compile(pattern)
		if err != nil {
			return err
		}
		f.regex = r
		return nil
	}

	if g := info.GetQuery().GetFilter().GetGauge(); g != nil {
		if g.Name == "" {
			return nil
		}

		for n, _ := range g.GetFilter() {
			if n == g.Name {
				return nil
			}
		}

		return fmt.Errorf("Filter map must include name")
	}

	return nil
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

func (f filter) filterViaGauge(info *v1.QueryInfo, e *loggregator.Envelope) bool {
	if info.GetFilter().GetGauge() == nil {
		return true
	}

	if e.GetGauge() == nil {
		return false
	}

	for name, value := range info.GetFilter().GetGauge().GetFilter() {
		if !f.containsKeyValue(name, value, e.GetGauge().GetMetrics()) {
			return false
		}
	}

	return true
}

func (f filter) containsKeyValue(name string, value *v1.GaugeFilterValue, m map[string]*loggregator.GaugeValue) bool {
	for n, v := range m {
		if name == n && (value == nil || value.GetValue() == v.GetValue()) {
			return true
		}
	}
	return false
}
