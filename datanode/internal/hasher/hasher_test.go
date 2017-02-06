package hasher_test

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"testing"

	v2 "github.com/apoydence/loggrebutterfly/api/loggregator/v2"
	"github.com/golang/protobuf/proto"

	"github.com/apoydence/loggrebutterfly/datanode/internal/hasher"
	"github.com/apoydence/onpar"
	. "github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

	os.Exit(m.Run())
}

type TH struct {
	*testing.T
	h *hasher.Hasher
}

func TestHasher(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TH {
		return TH{
			T: t,
			h: hasher.New(),
		}
	})

	o.Spec("returns the same hash for an envelope", func(t TH) {
		eA := &v2.Envelope{SourceUuid: "some-id"}
		dataA, err := proto.Marshal(eA)
		Expect(t, err == nil).To(BeTrue())

		eB := &v2.Envelope{SourceUuid: "some-id"}
		dataB, err := proto.Marshal(eB)
		Expect(t, err == nil).To(BeTrue())

		eC := &v2.Envelope{SourceUuid: "some-other-id"}
		dataC, err := proto.Marshal(eC)
		Expect(t, err == nil).To(BeTrue())

		hashA, err := t.h.Hash(dataA)
		Expect(t, err == nil).To(BeTrue())

		hashB, err := t.h.Hash(dataB)
		Expect(t, err == nil).To(BeTrue())

		hashC, err := t.h.Hash(dataC)
		Expect(t, err == nil).To(BeTrue())

		Expect(t, hashA).To(Equal(hashB))
		Expect(t, hashA).To(Not(Equal(hashC)))
	})

	o.Spec("returns an error for a non envelope", func(t TH) {
		_, err := t.h.Hash([]byte("not envelope"))
		Expect(t, err == nil).To(BeFalse())
	})
}
