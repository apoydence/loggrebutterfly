package server_test

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/apoydence/loggrebutterfly/analyst/internal/network/server"
	loggregator "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	v1 "github.com/apoydence/loggrebutterfly/api/v1"
	"github.com/apoydence/onpar"
	. "github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
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

func TestServer(t *testing.T) {
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
			resp, err := t.s.Query(context.Background(), &v1.QueryInfo{SourceId: "id"})
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
			t.s.Query(context.Background(), &v1.QueryInfo{SourceId: "id", TimeRange: &v1.TimeRange{
				Start: 99,
				End:   101,
			}})

			Expect(t, t.mockCalc.CalculateInput.Route).To(
				Chain(Receive(), Equal("id")),
			)
			Expect(t, t.mockCalc.CalculateInput.AlgName).To(
				Chain(Receive(), Equal("timerange")),
			)
		})

		o.Spec("it includes the request in the meta", func(t TS) {
			info := &v1.QueryInfo{SourceId: "id", TimeRange: &v1.TimeRange{
				Start: 99,
				End:   101,
			}}
			t.s.Query(context.Background(), info)

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
			_, err := t.s.Query(context.Background(), &v1.QueryInfo{SourceId: "id"})
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
