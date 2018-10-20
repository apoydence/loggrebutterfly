package server_test

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"testing"

	"github.com/poy/loggrebutterfly/analyst/internal/network/server"
	loggregator "github.com/poy/loggrebutterfly/api/loggregator/v2"
	v1 "github.com/poy/loggrebutterfly/api/v1"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
	"github.com/golang/protobuf/proto"

	"google.golang.org/grpc/grpclog"
)

//go:generate hel

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
		grpclog.SetLogger(log.New(ioutil.Discard, "", log.LstdFlags))
	}

	os.Exit(m.Run())
}

type TS struct {
	*testing.T

	mockCalc *mockCalculator
	s        *server.Server
}

func TestServerQuery(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TS {
		mockCalc := newMockCalculator()
		return TS{
			T:        t,
			mockCalc: mockCalc,
			s:        server.New(mockCalc),
		}
	})

	o.Group("when the calculator does not return an error", func() {
		o.BeforeEach(func(t TS) TS {
			close(t.mockCalc.CalculateOutput.Err)
			t.mockCalc.CalculateOutput.FinalResult <- map[string][]byte{
				"a":       marshalEnvelope("a"),
				"b":       marshalEnvelope("b"),
				"invalid": []byte("invalid"),
			}
			return t
		})

		o.Spec("it uses the calculator and returns the results", func(t TS) {
			resp, err := t.s.Query(context.Background(), &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "id",
				},
			})
			Expect(t, err == nil).To(BeTrue())

			Expect(t, resp.Envelopes).To(HaveLen(2))
			Expect(t, resp.Envelopes[0].SourceId).To(Or(
				Equal("a"),
				Equal("b"),
			))
			Expect(t, resp.Envelopes[1].SourceId).To(Or(
				Equal("a"),
				Equal("b"),
			))
			Expect(t, resp.Envelopes[0].SourceId).To(Not(Equal(resp.Envelopes[1].SourceId)))
		})

		o.Spec("it returns an error if an ID is not given", func(t TS) {
			_, err := t.s.Query(context.Background(), &v1.QueryInfo{})
			Expect(t, err == nil).To(BeFalse())
		})

		o.Spec("it uses the expected info for the calculator", func(t TS) {
			t.s.Query(context.Background(), &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "id", TimeRange: &v1.TimeRange{
						Start: 99,
						End:   101,
					},
				}})

			Expect(t, t.mockCalc.CalculateInput.Route).To(
				Chain(Receive(), Equal("id")),
			)
			Expect(t, t.mockCalc.CalculateInput.AlgName).To(
				Chain(Receive(), Equal("timerange")),
			)
		})

		o.Spec("it includes the request in the meta", func(t TS) {
			info := &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "id", TimeRange: &v1.TimeRange{
						Start: 99,
						End:   101,
					},
				}}
			t.s.Query(context.Background(), info)

			marshelled, err := proto.Marshal(&v1.AggregateInfo{Query: info})
			Expect(t, err == nil).To(BeTrue())

			Expect(t, t.mockCalc.CalculateInput.Meta).To(
				Chain(Receive(), Equal(marshelled)),
			)
		})
	})

	o.Group("when the calculator returns an error", func() {
		o.BeforeEach(func(t TS) TS {
			t.mockCalc.CalculateOutput.Err <- fmt.Errorf("some-error")
			close(t.mockCalc.CalculateOutput.FinalResult)
			return t
		})

		o.Spec("it returns an error", func(t TS) {
			_, err := t.s.Query(context.Background(), &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "id",
				},
			})
			Expect(t, err == nil).To(BeFalse())
		})
	})
}

func TestServerAggregate(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TS {
		mockCalc := newMockCalculator()
		return TS{
			T:        t,
			mockCalc: mockCalc,
			s:        server.New(mockCalc),
		}
	})

	o.Group("when the calculator does not return an error", func() {
		o.BeforeEach(func(t TS) TS {
			close(t.mockCalc.CalculateOutput.Err)
			t.mockCalc.CalculateOutput.FinalResult <- map[string][]byte{
				"0":       marshalFloat64(99),
				"1":       marshalFloat64(101),
				"2":       []byte("invalid"),
				"invalid": marshalFloat64(103),
			}
			return t
		})

		o.Spec("it uses the calculator and returns the results", func(t TS) {
			resp, err := t.s.Aggregate(context.Background(), &v1.AggregateInfo{
				Query: &v1.QueryInfo{
					Filter: &v1.AnalystFilter{
						SourceId: "some-id",
						Envelopes: &v1.AnalystFilter_Counter{
							Counter: &v1.CounterFilter{Name: "some-name"},
						},
					},
				},
				BucketWidthNs: 2,
			})
			Expect(t, err == nil).To(BeTrue())

			Expect(t, resp.Results).To(HaveLen(2))
			Expect(t, resp.Results[0]).To(Equal(float64(99)))
			Expect(t, resp.Results[1]).To(Equal(float64(101)))
		})

		o.Spec("it returns an error if an ID is not given", func(t TS) {
			_, err := t.s.Aggregate(context.Background(), &v1.AggregateInfo{
				Query: &v1.QueryInfo{
					Filter: &v1.AnalystFilter{
						Envelopes: &v1.AnalystFilter_Counter{
							Counter: &v1.CounterFilter{Name: "some-name"},
						},
					},
				},
				BucketWidthNs: 2,
			})
			Expect(t, err == nil).To(BeFalse())
		})

		o.Spec("it returns an error if an aggregation is not given", func(t TS) {
			_, err := t.s.Aggregate(context.Background(), &v1.AggregateInfo{
				BucketWidthNs: 2,
				Query: &v1.QueryInfo{
					Filter: &v1.AnalystFilter{
						SourceId: "some-id",
					},
				},
			})
			Expect(t, err == nil).To(BeFalse())
		})

		o.Spec("it returns an error if an bucket widtch is not given", func(t TS) {
			_, err := t.s.Aggregate(context.Background(), &v1.AggregateInfo{
				Query: &v1.QueryInfo{
					Filter: &v1.AnalystFilter{
						SourceId: "some-id",
						Envelopes: &v1.AnalystFilter_Counter{
							Counter: &v1.CounterFilter{Name: "some-name"},
						},
					},
				},
			})
			Expect(t, err == nil).To(BeFalse())
		})

		o.Spec("it uses the expected info for the calculator", func(t TS) {
			t.s.Aggregate(context.Background(), &v1.AggregateInfo{Query: &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "id",
					TimeRange: &v1.TimeRange{
						Start: 99,
						End:   101,
					},
					Envelopes: &v1.AnalystFilter_Counter{
						Counter: &v1.CounterFilter{Name: "some-name"},
					},
				},
			},
				BucketWidthNs: 2,
			})

			Expect(t, t.mockCalc.CalculateInput.Route).To(
				Chain(Receive(), Equal("id")),
			)
			Expect(t, t.mockCalc.CalculateInput.AlgName).To(
				Chain(Receive(), Equal("aggregation")),
			)
		})

		o.Spec("it includes the request in the meta", func(t TS) {
			info := &v1.AggregateInfo{Query: &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "id",
					TimeRange: &v1.TimeRange{
						Start: 99,
						End:   101,
					},
					Envelopes: &v1.AnalystFilter_Counter{
						Counter: &v1.CounterFilter{Name: "some-name"},
					},
				},
			},
				BucketWidthNs: 2,
			}
			t.s.Aggregate(context.Background(), info)

			marshelled, err := proto.Marshal(info)
			Expect(t, err == nil).To(BeTrue())

			Expect(t, t.mockCalc.CalculateInput.Meta).To(
				Chain(Receive(), Equal(marshelled)),
			)
		})
	})

	o.Group("when the calculator returns an error", func() {
		o.BeforeEach(func(t TS) TS {
			t.mockCalc.CalculateOutput.Err <- fmt.Errorf("some-error")
			close(t.mockCalc.CalculateOutput.FinalResult)
			return t
		})

		o.Spec("it returns an error", func(t TS) {
			_, err := t.s.Aggregate(context.Background(), &v1.AggregateInfo{Query: &v1.QueryInfo{
				Filter: &v1.AnalystFilter{
					SourceId: "id",
					TimeRange: &v1.TimeRange{
						Start: 99,
						End:   101,
					},
					Envelopes: &v1.AnalystFilter_Counter{
						Counter: &v1.CounterFilter{Name: "some-name"},
					},
				},
			},
				BucketWidthNs: 2,
			})
			Expect(t, err == nil).To(BeFalse())
		})
	})
}

func marshalEnvelope(sourceId string) []byte {
	e := &loggregator.Envelope{SourceId: sourceId}
	data, err := proto.Marshal(e)
	if err != nil {
		panic(err)
	}
	return data
}

func marshalFloat64(f float64) []byte {
	bits := math.Float64bits(f)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}
