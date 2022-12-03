package source

import (
	"bufio"
	"io"
	"os"
	"sync/atomic"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type FileSourceOptions struct {
	FilePath         string
	IgnoreParseError bool
}

func NewFileSourceWithOptions(options *FileSourceOptions) (*FileSource, error) {
	fp, err := os.Open(options.FilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "os.Open [%s] failed", options.FilePath)
	}

	var source []interface{}
	reader := bufio.NewReader(fp)
	for {
		buf, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return nil, errors.Wrapf(err, "reader.ReadString failed")
		}
		if len(buf) == 1 {
			continue
		}
		var v interface{}
		if err := jsoniter.Unmarshal(buf, &v); err != nil {
			if options.IgnoreParseError {
				continue
			}
			return nil, errors.Wrapf(err, "parse [%s] failed. file [%s]", string(buf), options.FilePath)
		}

		source = append(source, v)

		if err != nil {
			break
		}
	}

	return &FileSource{
		source: source,
		len:    uint64(len(source)),
	}, nil
}

type FileSource struct {
	source []interface{}
	idx    uint64
	len    uint64
}

func (s *FileSource) Fetch() interface{} {
	idx := atomic.AddUint64(&s.idx, 1)
	idx %= s.len

	return s.source[idx]
}
