package mappers_test

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/apoydence/loggrebutterfly/analyst/internal/algorithms/mappers"
	loggregator "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	v1 "github.com/apoydence/loggrebutterfly/api/v1"
	"github.com/apoydence/onpar"
	. "github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
)

type TA struct {
	*testing.T
	agg mappers.Aggregation
}

func TestAggregation(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

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
		return TA{
			T:   t,
			agg: mappers.NewAggregation(req),
		}
	})

	o.Spec("it only returns envelopes that have the correct source ID", func(t TA) {
		e1 := buildCounter("some-name", "wrong", 98)
		e2 := buildCounter("some-name", "some-id", 99)

		key, _, err := t.agg.Map(e1)
		Expect(t, err == nil).To(BeTrue())
		Expect(t, key).To(HaveLen(0))

		key, _, err = t.agg.Map(e2)
		Expect(t, err == nil).To(BeTrue())
		Expect(t, key).To(Not(HaveLen(0)))
	})

	o.Spec("it filters out envelopes that are outside the time range or the wrong name", func(t TA) {
		e1 := buildCounter("some-name", "some-id", 98)
		e2 := buildCounter("some-name", "some-id", 99)
		e3 := buildCounter("some-name", "some-id", 100)
		e4 := buildCounter("some-name", "some-id", 101)
		e5 := buildCounter("wrong", "some-id", 99)
		e6 := marshalEnvelope(&loggregator.Envelope{SourceId: "some-id", Timestamp: 99})

		key, _, _ := t.agg.Map(e1)
		Expect(t, key).To(HaveLen(0))

		key, _, _ = t.agg.Map(e2)
		Expect(t, key).To(Equal("98"))

		key, _, _ = t.agg.Map(e3)
		Expect(t, key).To(Equal("100"))

		key, _, _ = t.agg.Map(e4)
		Expect(t, key).To(HaveLen(0))

		key, _, _ = t.agg.Map(e5)
		Expect(t, key).To(HaveLen(0))

		key, _, _ = t.agg.Map(e6)
		Expect(t, key).To(HaveLen(0))
	})

	o.Spec("it uses the truncated timestamp as a key", func(t TA) {
		e := buildCounter("some-name", "some-id", 99)
		key, _, _ := t.agg.Map(e)
		Expect(t, key).To(Equal("98"))
	})

	o.Spec("it returns a float64 as a value", func(t TA) {
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
