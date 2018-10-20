package algorithms_test

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/poy/loggrebutterfly/analyst/internal/algorithms"
	v1 "github.com/poy/loggrebutterfly/api/v1"
	"github.com/poy/mapreduce"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
	"github.com/golang/protobuf/proto"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

	os.Exit(m.Run())
}

type TF struct {
	*testing.T
	f       *algorithms.Fetcher
	builder algorithms.AlgBuilder
}

func TestFetcher(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TF {
		builder := func(*v1.AggregateInfo) (mapreduce.Algorithm, error) { return mapreduce.Algorithm{}, nil }
		builders := map[string]algorithms.AlgBuilder{
			"a": builder,
		}

		return TF{
			T:       t,
			f:       algorithms.NewFetcher(builders),
			builder: builder,
		}
	})

	o.Spec("it returns the expected builder", func(t TF) {
		q := new(v1.QueryInfo)
		data, err := proto.Marshal(q)
		Expect(t, err == nil).To(BeTrue())

		builder, err := t.f.Alg("a", data)
		Expect(t, err == nil).To(BeTrue())
		Expect(t, builder).To(Equal(mapreduce.Algorithm{}))
	})

	o.Spec("it returns an error for an unknown alg", func(t TF) {
		q := new(v1.QueryInfo)
		data, err := proto.Marshal(q)
		Expect(t, err == nil).To(BeTrue())

		_, err = t.f.Alg("invalid", data)
		Expect(t, err == nil).To(BeFalse())
	})

	o.Spec("it returns an error if the context does not have a valid request", func(t TF) {
		_, err := t.f.Alg("a", []byte("invalid"))
		Expect(t, err == nil).To(BeFalse())
	})
}
