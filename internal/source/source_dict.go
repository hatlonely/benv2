package source

import "sync/atomic"

type DictSourceOptions []interface{}

func NewDictSourceWithOptions(options *DictSourceOptions) *DictSource {
	return &DictSource{
		source: *options,
		len:    uint64(len(*options)),
	}
}

type DictSource struct {
	source []interface{}
	idx    uint64
	len    uint64
}

func (s *DictSource) Fetch() interface{} {
	idx := atomic.AddUint64(&s.idx, 1)
	idx %= s.len

	return s.source[idx]
}
