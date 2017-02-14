package reducers_test

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/apoydence/loggrebutterfly/analyst/internal/algorithms/reducers"
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

type TF struct {
	*testing.T
	f reducers.First
}

func TestFirst(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TF {
		return TF{
			T: t,
			f: reducers.NewFirst(),
		}
	})

	o.Spec("it returns the first entry", func(t TF) {
		result, err := t.f.Reduce([][]byte{[]byte("a"), []byte("b")})
		Expect(t, err == nil).To(BeTrue())
		Expect(t, result).To(HaveLen(1))
		Expect(t, result[0]).To(Equal([]byte("a")))
	})

	o.Spec("it returns an empty list for an empty list", func(t TF) {
		result, err := t.f.Reduce(nil)
		Expect(t, err == nil).To(BeTrue())
		Expect(t, result).To(HaveLen(0))
	})
}
