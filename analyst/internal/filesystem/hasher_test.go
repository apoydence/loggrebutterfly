package filesystem_test

import (
	"testing"

	"github.com/poy/loggrebutterfly/analyst/internal/filesystem"

	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
)

type TH struct {
	*testing.T
	h *filesystem.StringHasher
}

func TestHasher(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TH {
		return TH{
			T: t,
			h: filesystem.NewHasher(),
		}
	})

	o.Spec("returns the same hash for a string", func(t TH) {
		hashA := t.h.HashString("some-id")
		hashB := t.h.HashString("some-id")
		hashC := t.h.HashString("some-other-id")

		Expect(t, hashA).To(Equal(hashB))
		Expect(t, hashA).To(Not(Equal(hashC)))
	})
}
