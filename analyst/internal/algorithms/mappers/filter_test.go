package mappers_test

import (
	"testing"

	"github.com/apoydence/loggrebutterfly/analyst/internal/algorithms/mappers"
	loggregator "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	v1 "github.com/apoydence/loggrebutterfly/api/v1"
	"github.com/apoydence/onpar"
	. "github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
)

type TF struct {
	*testing.T
	tr mappers.Filter
}

func TestFilter(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.Group("timerange", func() {
		o.BeforeEach(func(t *testing.T) TF {
			req := &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "some-id",
					TimeRange: &v1.TimeRange{
						Start: 99,
						End:   101,
					},
				},
			}

			return TF{
				T:  t,
				tr: mappers.NewFilter(&v1.AggregateInfo{Query: req}),
			}
		})

		o.Spec("it only returns envelopes that have the correct source ID", func(t TF) {
			e1 := &loggregator.Envelope{SourceId: "wrong"}
			e2 := &loggregator.Envelope{SourceId: "some-id", Timestamp: 99}

			keep := t.tr.Filter(e1)
			Expect(t, keep).To(BeFalse())

			keep = t.tr.Filter(e2)
			Expect(t, keep).To(BeTrue())
		})

		o.Spec("it filters out envelopes that are outside the time range", func(t TF) {
			e1 := &loggregator.Envelope{SourceId: "some-id", Timestamp: 98}
			e2 := &loggregator.Envelope{SourceId: "some-id", Timestamp: 99}
			e3 := &loggregator.Envelope{SourceId: "some-id", Timestamp: 100}
			e4 := &loggregator.Envelope{SourceId: "some-id", Timestamp: 101}

			keep := t.tr.Filter(e1)
			Expect(t, keep).To(BeFalse())

			keep = t.tr.Filter(e2)
			Expect(t, keep).To(BeTrue())

			keep = t.tr.Filter(e3)
			Expect(t, keep).To(BeTrue())

			keep = t.tr.Filter(e4)
			Expect(t, keep).To(BeFalse())
		})
	})

	o.Group("CounterFilter", func() {
		o.Group("empty name", func() {
			o.BeforeEach(func(t *testing.T) TF {
				req := &v1.QueryInfo{
					Filter: &v1.AnalystFilter{
						SourceId: "some-id",
						Envelopes: &v1.AnalystFilter_Counter{
							Counter: &v1.CounterFilter{},
						},
					},
				}

				return TF{
					T:  t,
					tr: mappers.NewFilter(&v1.AggregateInfo{Query: req}),
				}
			})

			o.Spec("it filters out envelopes that are not counters", func(t TF) {
				e1 := &loggregator.Envelope{SourceId: "some-id", Timestamp: 98}
				e2 := &loggregator.Envelope{
					SourceId:  "some-id",
					Timestamp: 98,
					Message: &loggregator.Envelope_Counter{
						Counter: &loggregator.Counter{
							Name: "some-name",
						},
					},
				}

				keep := t.tr.Filter(e1)
				Expect(t, keep).To(BeFalse())

				keep = t.tr.Filter(e2)
				Expect(t, keep).To(BeTrue())
			})
		})

		o.Group("non-empty name", func() {
			o.BeforeEach(func(t *testing.T) TF {
				req := &v1.QueryInfo{
					Filter: &v1.AnalystFilter{
						SourceId: "some-id",
						Envelopes: &v1.AnalystFilter_Counter{
							Counter: &v1.CounterFilter{
								Name: "some-name",
							},
						},
					},
				}

				return TF{
					T:  t,
					tr: mappers.NewFilter(&v1.AggregateInfo{Query: req}),
				}
			})

			o.Spec("it filters out envelopes that are not the right name", func(t TF) {
				e1 := &loggregator.Envelope{
					SourceId:  "some-id",
					Timestamp: 97,
					Message: &loggregator.Envelope_Counter{
						Counter: &loggregator.Counter{
							Name: "wrong-name",
						},
					},
				}
				e2 := &loggregator.Envelope{
					SourceId:  "some-id",
					Timestamp: 98,
					Message: &loggregator.Envelope_Counter{
						Counter: &loggregator.Counter{
							Name: "some-name",
						},
					},
				}

				keep := t.tr.Filter(e1)
				Expect(t, keep).To(BeFalse())

				keep = t.tr.Filter(e2)
				Expect(t, keep).To(BeTrue())
			})
		})
	})
}
