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

func TestFilterTimerange(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

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

		f, err := mappers.NewFilter(&v1.AggregateInfo{Query: req})
		Expect(t, err == nil).To(BeTrue())

		return TF{
			T:  t,
			tr: f,
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
}

func TestFilterLogFilter(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.Group("Empty payload", func() {
		o.BeforeEach(func(t *testing.T) TF {
			req := &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "some-id",
					Envelopes: &v1.AnalystFilter_Log{
						Log: &v1.LogFilter{},
					},
					TimeRange: &v1.TimeRange{
						Start: 1,
						End:   9223372036854775807,
					},
				},
			}

			f, err := mappers.NewFilter(&v1.AggregateInfo{Query: req})
			Expect(t, err == nil).To(BeTrue())

			return TF{
				T:  t,
				tr: f,
			}
		})

		o.Spec("it filters out envelopes that are not logs", func(t TF) {
			e1 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 98,
				Message: &loggregator.Envelope_Counter{
					Counter: &loggregator.Counter{
						Name: "some-name",
					},
				},
			}

			e2 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 98,
				Message: &loggregator.Envelope_Log{
					Log: &loggregator.Log{
						Payload: []byte("some-value"),
					},
				},
			}

			keep := t.tr.Filter(e1)
			Expect(t, keep).To(BeFalse())

			keep = t.tr.Filter(e2)
			Expect(t, keep).To(BeTrue())
		})
	})

	o.Group("Match", func() {
		o.BeforeEach(func(t *testing.T) TF {
			req := &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "some-id",
					Envelopes: &v1.AnalystFilter_Log{
						Log: &v1.LogFilter{
							Payload: &v1.LogFilter_Match{
								Match: []byte("some-value"),
							},
						},
					},
				},
			}

			f, err := mappers.NewFilter(&v1.AggregateInfo{Query: req})
			Expect(t, err == nil).To(BeTrue())

			return TF{
				T:  t,
				tr: f,
			}
		})

		o.Spec("it filters out envelopes that are not logs", func(t TF) {
			e1 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 98,
				Message: &loggregator.Envelope_Counter{
					Counter: &loggregator.Counter{
						Name: "some-name",
					},
				},
			}

			e2 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 98,
				Message: &loggregator.Envelope_Log{
					Log: &loggregator.Log{
						Payload: []byte("some-value"),
					},
				},
			}

			keep := t.tr.Filter(e1)
			Expect(t, keep).To(BeFalse())

			keep = t.tr.Filter(e2)
			Expect(t, keep).To(BeTrue())
		})

		o.Spec("it filters out envelopes that don't have the exact payload", func(t TF) {
			e1 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 98,
				Message: &loggregator.Envelope_Log{
					Log: &loggregator.Log{
						Payload: []byte("wrong-value"),
					},
				},
			}

			e2 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 98,
				Message: &loggregator.Envelope_Log{
					Log: &loggregator.Log{
						Payload: []byte("some-value"),
					},
				},
			}

			keep := t.tr.Filter(e1)
			Expect(t, keep).To(BeFalse())

			keep = t.tr.Filter(e2)
			Expect(t, keep).To(BeTrue())
		})
	})

	o.Group("Regexp", func() {
		o.BeforeEach(func(t *testing.T) TF {
			req := &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "some-id",
					Envelopes: &v1.AnalystFilter_Log{
						Log: &v1.LogFilter{
							Payload: &v1.LogFilter_Regexp{
								Regexp: "^[sS]ome-value",
							},
						},
					},
				},
			}

			f, err := mappers.NewFilter(&v1.AggregateInfo{Query: req})
			Expect(t, err == nil).To(BeTrue())

			return TF{
				T:  t,
				tr: f,
			}
		})

		o.Spec("it returns an error for an invalid regexp pattern", func(t TF) {
			req := &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "some-id",
					Envelopes: &v1.AnalystFilter_Log{
						Log: &v1.LogFilter{
							Payload: &v1.LogFilter_Regexp{
								Regexp: "[",
							},
						},
					},
				},
			}

			_, err := mappers.NewFilter(&v1.AggregateInfo{Query: req})
			Expect(t, err == nil).To(BeFalse())
		})

		o.Spec("it filters out envelopes that are not logs", func(t TF) {
			e1 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 98,
				Message: &loggregator.Envelope_Counter{
					Counter: &loggregator.Counter{
						Name: "some-name",
					},
				},
			}

			e2 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 98,
				Message: &loggregator.Envelope_Log{
					Log: &loggregator.Log{
						Payload: []byte("some-value"),
					},
				},
			}

			keep := t.tr.Filter(e1)
			Expect(t, keep).To(BeFalse())

			keep = t.tr.Filter(e2)
			Expect(t, keep).To(BeTrue())
		})

		o.Spec("it filters out envelopes that do not match the regexp", func(t TF) {
			e1 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 98,
				Message: &loggregator.Envelope_Log{
					Log: &loggregator.Log{
						Payload: []byte("wrong-some-value"),
					},
				},
			}

			e2 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 98,
				Message: &loggregator.Envelope_Log{
					Log: &loggregator.Log{
						Payload: []byte("some-value-thats-good"),
					},
				},
			}

			keep := t.tr.Filter(e1)
			Expect(t, keep).To(BeFalse())

			keep = t.tr.Filter(e2)
			Expect(t, keep).To(BeTrue())
		})
	})
}

func TestFilterCounter(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

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

			f, err := mappers.NewFilter(&v1.AggregateInfo{Query: req})
			Expect(t, err == nil).To(BeTrue())

			return TF{
				T:  t,
				tr: f,
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

			f, err := mappers.NewFilter(&v1.AggregateInfo{Query: req})
			Expect(t, err == nil).To(BeTrue())

			return TF{
				T:  t,
				tr: f,
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
}

func TestFilterGauge(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.Group("empty names", func() {
		o.BeforeEach(func(t *testing.T) TF {
			req := &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "some-id",
					Envelopes: &v1.AnalystFilter_Gauge{
						Gauge: &v1.GaugeFilter{},
					},
				},
			}

			f, err := mappers.NewFilter(&v1.AggregateInfo{Query: req})
			Expect(t, err == nil).To(BeTrue())

			return TF{
				T:  t,
				tr: f,
			}
		})

		o.Spec("it filters out envelopes that are not gauges", func(t TF) {
			e1 := &loggregator.Envelope{SourceId: "some-id", Timestamp: 98}
			e2 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 98,
				Message: &loggregator.Envelope_Gauge{
					Gauge: &loggregator.Gauge{
						Metrics: map[string]*loggregator.GaugeValue{
							"some-name": &loggregator.GaugeValue{
								Value: 5.5,
							},
						},
					},
				},
			}

			keep := t.tr.Filter(e1)
			Expect(t, keep).To(BeFalse())

			keep = t.tr.Filter(e2)
			Expect(t, keep).To(BeTrue())
		})
	})

	o.Group("multiple names without values", func() {
		o.BeforeEach(func(t *testing.T) TF {
			req := &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "some-id",
					Envelopes: &v1.AnalystFilter_Gauge{
						Gauge: &v1.GaugeFilter{
							Filter: map[string]*v1.GaugeFilterValue{
								"a": nil,
								"b": &v1.GaugeFilterValue{99},
							},
						},
					},
				},
			}

			f, err := mappers.NewFilter(&v1.AggregateInfo{Query: req})
			Expect(t, err == nil).To(BeTrue())

			return TF{
				T:  t,
				tr: f,
			}
		})

		o.Spec("it filters out envelopes that have a key value mismatch", func(t TF) {
			e1 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 97,
				Message: &loggregator.Envelope_Gauge{
					Gauge: &loggregator.Gauge{
						Metrics: map[string]*loggregator.GaugeValue{
							"some-name": &loggregator.GaugeValue{
								Value: 5.5,
							},
						},
					},
				},
			}
			e2 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 97,
				Message: &loggregator.Envelope_Gauge{
					Gauge: &loggregator.Gauge{
						Metrics: map[string]*loggregator.GaugeValue{
							"some-name": &loggregator.GaugeValue{
								Value: 5.5,
							},
							"a": &loggregator.GaugeValue{
								Value: 5.5,
							},
							"b": &loggregator.GaugeValue{
								Value: 5.5,
							},
						},
					},
				},
			}
			e3 := &loggregator.Envelope{
				SourceId:  "some-id",
				Timestamp: 97,
				Message: &loggregator.Envelope_Gauge{
					Gauge: &loggregator.Gauge{
						Metrics: map[string]*loggregator.GaugeValue{
							"some-name": &loggregator.GaugeValue{
								Value: 5.5,
							},
							"a": &loggregator.GaugeValue{
								Value: 5.5,
							},
							"b": &loggregator.GaugeValue{
								Value: 99,
							},
						},
					},
				},
			}

			keep := t.tr.Filter(e1)
			Expect(t, keep).To(BeFalse())

			keep = t.tr.Filter(e2)
			Expect(t, keep).To(BeFalse())

			keep = t.tr.Filter(e3)
			Expect(t, keep).To(BeTrue())
		})
	})
}
