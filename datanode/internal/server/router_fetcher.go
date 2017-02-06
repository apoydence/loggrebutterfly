package server

import "github.com/apoydence/petasos/router"

type RouterFetcher struct {
	fs             router.FileSystem
	hasher         router.Hasher
	metricsCounter router.MetricsCounter
}

func NewRouterFetcher(
	fs router.FileSystem,
	hasher router.Hasher,
	metricsCounter router.MetricsCounter,
) *RouterFetcher {
	return &RouterFetcher{fs: fs, hasher: hasher, metricsCounter: metricsCounter}
}

func (f *RouterFetcher) Writer() (writer func(data []byte) (err error), err error) {
	return router.New(f.fs, f.hasher, f.metricsCounter).Write, nil
}
