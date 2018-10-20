package mappers_test

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/poy/loggrebutterfly/analyst/internal/algorithms/mappers"
	loggregator "github.com/poy/loggrebutterfly/api/loggregator/v2"
	v1 "github.com/poy/loggrebutterfly/api/v1"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
)

type TA struct {
	*testing.T
	mockFilter *mockFilter
	agg        mappers.Aggregation
}

func TestAggregation(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.Group("counter", func() {
		o.BeforeEach(func(t *testing.T) TA {
			req := &v1.AggregateInfo{
				BucketWidthNs: 2,
				Query: &v1.QueryInfo{
					Filter: &v1.AnalystFilter{
						SourceId: "some-id",
						TimeRange: &v1.TimeRange{
							Start: 99,
							End:   101,
						},
						Envelopes: &v1.AnalystFilter_Counter{
							Counter: &v1.CounterFilter{
								Name: "some-name",
							},
						},
					},
				},
			}

			mockFilter := newMockFilter()
			agg, err := mappers.NewAggregation(req, mockFilter)
			Expect(t, err == nil).To(BeTrue())
			return TA{
				T:          t,
				mockFilter: mockFilter,
				agg:        agg,
			}
		})

		o.Spec("it uses the truncated timestamp as a key", func(t TA) {
			t.mockFilter.FilterOutput.Keep <- true
			e := buildCounter("some-name", "some-id", 99)
			key, _, _ := t.agg.Map(e)
			Expect(t, key).To(Equal("98"))
		})

		o.Spec("it uses an empty key for filtered out envelopes", func(t TA) {
			t.mockFilter.FilterOutput.Keep <- false
			e := marshalEnvelope(&loggregator.Envelope{SourceId: "some-id", Timestamp: 99})
			key, _, _ := t.agg.Map(e)
			Expect(t, key).To(HaveLen(0))
		})

		o.Spec("it returns a float64 as a value", func(t TA) {
			t.mockFilter.FilterOutput.Keep <- true
			e := buildCounter("some-name", "some-id", 99)
			_, value, _ := t.agg.Map(e)
			bits := binary.LittleEndian.Uint64(value)
			float := math.Float64frombits(bits)

			Expect(t, float).To(Equal(float64(999)))
		})

		o.Spec("it returns an error for a non-envelope", func(t TA) {
			_, _, err := t.agg.Map([]byte("invalid"))
			Expect(t, err == nil).To(BeFalse())
		})
	})

	o.Group("gauge", func() {
		o.BeforeEach(func(t *testing.T) TA {
			req := &v1.AggregateInfo{
				BucketWidthNs: 2,
				Query: &v1.QueryInfo{
					Filter: &v1.AnalystFilter{
						SourceId: "some-id",
						TimeRange: &v1.TimeRange{
							Start: 99,
							End:   101,
						},
						Envelopes: &v1.AnalystFilter_Gauge{
							Gauge: &v1.GaugeFilter{
								Name: "some-name",
							},
						},
					},
				},
			}

			mockFilter := newMockFilter()
			agg, err := mappers.NewAggregation(req, mockFilter)
			Expect(t, err == nil).To(BeTrue())
			return TA{
				T:          t,
				mockFilter: mockFilter,
				agg:        agg,
			}
		})

		o.Spec("it returns the value for the given name", func(t TA) {
			t.mockFilter.FilterOutput.Keep <- true
			e := buildGauge("some-name", "some-id", 99)
			_, value, _ := t.agg.Map(e)
			bits := binary.LittleEndian.Uint64(value)
			float := math.Float64frombits(bits)

			Expect(t, float).To(Equal(float64(888)))
		})
	})

}

func TestAggregationInvalidFilter(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	req := &v1.AggregateInfo{
		BucketWidthNs: 2,
		Query: &v1.QueryInfo{
			Filter: &v1.AnalystFilter{
				SourceId: "some-id",
				TimeRange: &v1.TimeRange{
					Start: 99,
					End:   101,
				},
				Envelopes: &v1.AnalystFilter_Log{
					Log: &v1.LogFilter{
						Payload: &v1.LogFilter_Match{
							Match: []byte("something"),
						},
					},
				},
			},
		},
	}

	mockFilter := newMockFilter()
	_, err := mappers.NewAggregation(req, mockFilter)
	Expect(t, err == nil).To(BeFalse())
}

func buildCounter(name, sourceId string, t int64) []byte {
	return marshalEnvelope(&loggregator.Envelope{
		SourceId:  sourceId,
		Timestamp: t,
		Message: &loggregator.Envelope_Counter{
			Counter: &loggregator.Counter{
				Name: name,
				Value: &loggregator.Counter_Total{
					Total: 999,
				},
			},
		},
	})
}

func buildGauge(name, sourceId string, t int64) []byte {
	return marshalEnvelope(&loggregator.Envelope{
		SourceId:  sourceId,
		Timestamp: t,
		Message: &loggregator.Envelope_Gauge{
			Gauge: &loggregator.Gauge{
				Metrics: map[string]*loggregator.GaugeValue{
					name:         &loggregator.GaugeValue{Value: 888},
					"other-name": &loggregator.GaugeValue{Value: 777},
				},
			},
		},
	})
}
