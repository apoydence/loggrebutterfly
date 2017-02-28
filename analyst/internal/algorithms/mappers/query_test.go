package mappers_test

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/apoydence/loggrebutterfly/analyst/internal/algorithms/mappers"
	loggregator "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	"github.com/apoydence/onpar"
	. "github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
	"github.com/golang/protobuf/proto"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

	os.Exit(m.Run())
}

type TTR struct {
	*testing.T
	mockFilter *mockFilter
	tr         mappers.Query
}

func TestQuery(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.Group("timerange", func() {
		o.BeforeEach(func(t *testing.T) TTR {
			mockFilter := newMockFilter()

			return TTR{
				T:          t,
				tr:         mappers.NewQuery(mockFilter),
				mockFilter: mockFilter,
			}
		})

		o.Spec("it uses the timestamp as a key", func(t TTR) {
			t.mockFilter.FilterOutput.Keep <- true
			e := marshalEnvelope(&loggregator.Envelope{SourceId: "some-id", Timestamp: 99})
			key, _, _ := t.tr.Map(e)
			Expect(t, key).To(Equal("99"))
		})

		o.Spec("it uses an empty key for filtered out envelopes", func(t TTR) {
			t.mockFilter.FilterOutput.Keep <- false
			e := marshalEnvelope(&loggregator.Envelope{SourceId: "some-id", Timestamp: 99})
			key, _, _ := t.tr.Map(e)
			Expect(t, key).To(HaveLen(0))
		})

		o.Spec("it returns an error for a non-envelope", func(t TTR) {
			_, _, err := t.tr.Map([]byte("invalid"))
			Expect(t, err == nil).To(BeFalse())
		})
	})
}

func marshalEnvelope(e *loggregator.Envelope) []byte {
	d, err := proto.Marshal(e)
	if err != nil {
		panic(err)
	}

	return d
}
