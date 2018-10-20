package reducers_test

import (
	"encoding/binary"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"os"
	"testing"

	"github.com/poy/loggrebutterfly/analyst/internal/algorithms/reducers"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
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
	s reducers.SumF
}

func TestFirst(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TF {
		return TF{
			T: t,
			f: reducers.NewFirst(),
			s: reducers.NewSumF(),
		}
	})

	o.Group("First", func() {
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
	})

	o.Group("SumF", func() {
		o.Spec("it returns the sum of all the values as float64s", func(t TF) {
			result, err := t.s.Reduce([][]byte{
				floatToBytes(1),
				floatToBytes(2),
				floatToBytes(3),
			})
			Expect(t, err == nil).To(BeTrue())
			Expect(t, result).To(HaveLen(1))
			Expect(t, bytesToFloat(result[0])).To(Equal(float64(6)))
		})

		o.Spec("it returns an error for a non float64 value", func(t TF) {
			_, err := t.s.Reduce([][]byte{
				floatToBytes(1),
				floatToBytes(2),
				floatToBytes(3),
				[]byte("invalid"),
			})
			Expect(t, err == nil).To(BeFalse())
		})
	})
}

func floatToBytes(f float64) []byte {
	bits := math.Float64bits(f)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

func bytesToFloat(b []byte) float64 {
	bits := binary.LittleEndian.Uint64(b)
	float := math.Float64frombits(bits)
	return float
}
