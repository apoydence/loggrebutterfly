package mappers_test

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/apoydence/loggrebutterfly/analyst/internal/algorithms/mappers"
	loggregator "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	v1 "github.com/apoydence/loggrebutterfly/api/v1"
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
	tr mappers.TimeRange
}

func TestTimerange(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TTR {
		req := &v1.QueryInfo{
			SourceId: "some-id",
			TimeRange: &v1.TimeRange{
				Start: 99,
				End:   101,
			},
		}
		return TTR{
			T:  t,
			tr: mappers.NewTimeRange(req),
		}
	})

	o.Spec("it only returns envelopes that have the correct source ID", func(t TTR) {
		e1 := marshalEnvelope(&loggregator.Envelope{SourceId: "wrong"})
		e2 := marshalEnvelope(&loggregator.Envelope{SourceId: "some-id", Timestamp: 99})

		key, _, err := t.tr.Map(e1)
		Expect(t, err == nil).To(BeTrue())
		Expect(t, key).To(HaveLen(0))

		key, value, err := t.tr.Map(e2)
		Expect(t, err == nil).To(BeTrue())
		Expect(t, key).To(Not(HaveLen(0)))
		Expect(t, value).To(Equal(e2))
	})

	o.Spec("it filters out envelopes that are outside the time range", func(t TTR) {
		e1 := marshalEnvelope(&loggregator.Envelope{SourceId: "some-id", Timestamp: 98})
		e2 := marshalEnvelope(&loggregator.Envelope{SourceId: "some-id", Timestamp: 99})
		e3 := marshalEnvelope(&loggregator.Envelope{SourceId: "some-id", Timestamp: 100})
		e4 := marshalEnvelope(&loggregator.Envelope{SourceId: "some-id", Timestamp: 101})

		key, _, _ := t.tr.Map(e1)
		Expect(t, key).To(HaveLen(0))

		key, _, _ = t.tr.Map(e2)
		Expect(t, key).To(Equal("99"))

		key, _, _ = t.tr.Map(e3)
		Expect(t, key).To(Equal("100"))

		key, _, _ = t.tr.Map(e4)
		Expect(t, key).To(HaveLen(0))
	})

	o.Spec("it uses the timestamp as a key", func(t TTR) {
		e := marshalEnvelope(&loggregator.Envelope{SourceId: "some-id", Timestamp: 99})
		key, _, _ := t.tr.Map(e)
		Expect(t, key).To(Equal("99"))
	})

	o.Spec("it returns an error for a non-envelope", func(t TTR) {
		_, _, err := t.tr.Map([]byte("invalid"))
		Expect(t, err == nil).To(BeFalse())
	})
}

func marshalEnvelope(e *loggregator.Envelope) []byte {
	d, err := proto.Marshal(e)
	if err != nil {
		panic(err)
	}

	return d
}
