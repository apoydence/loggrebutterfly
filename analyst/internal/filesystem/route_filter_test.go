package filesystem_test

import (
	"testing"

	"github.com/poy/eachers/testhelpers"
	"github.com/poy/loggrebutterfly/analyst/internal/filesystem"
	"github.com/poy/onpar"
	. "github.com/poy/onpar/expect"
	. "github.com/poy/onpar/matchers"
)

type TRF struct {
	*testing.T
	f          *filesystem.RouteFilter
	mockHasher *mockHasher
}

func TestRouteFilter(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TRF {
		mockHasher := newMockHasher()
		testhelpers.AlwaysReturn(mockHasher.HashStringOutput.Hash, 99)

		return TRF{
			T:          t,
			mockHasher: mockHasher,
			f:          filesystem.NewRouteFilter(mockHasher),
		}
	})

	o.Spec("it prunes out unrelated routes", func(t TRF) {
		files := map[string][]string{
			"invalid":                 nil,
			`{"low":0, "high":99}`:    nil,
			`{"low":100, "high":199}`: nil,
			`{"low":99, "high":199}`:  nil,
			`{"low":0, "high":199}`:   nil,
		}

		t.f.Filter("some-route", files)
		Expect(t, files).To(HaveLen(3))
		Expect(t, files).To(HaveKey(`{"low":0, "high":99}`))
		Expect(t, files).To(HaveKey(`{"low":99, "high":199}`))
		Expect(t, files).To(HaveKey(`{"low":0, "high":199}`))
	})

}
