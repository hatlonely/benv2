package recorder

import (
	"bufio"
	"os"
	"sync"

	"github.com/hatlonely/go-kit/strx"
	"github.com/pkg/errors"
)

type FileRecorderOptions struct {
	FilePath string
	BufSize  int `dft:"32768"`
}

func NewFileRecorderWithOptions(options *FileRecorderOptions) (*FileRecorder, error) {
	fp, err := os.Create(options.FilePath)
	if err != nil {
		return nil, errors.WithMessage(err, "os.Create failed")
	}
	writer := bufio.NewWriterSize(fp, options.BufSize)

	return &FileRecorder{
		fp:     fp,
		writer: writer,
	}, nil
}

type FileRecorder struct {
	fp     *os.File
	writer *bufio.Writer
	mutex  sync.Mutex
}

func (r *FileRecorder) Close() error {
	if err := r.writer.Flush(); err != nil {
		return errors.Wrap(err, "writer.Flush failed")
	}
	if err := r.fp.Close(); err != nil {
		return errors.Wrap(err, "fp.Close failed")
	}
	return nil
}

func (r *FileRecorder) Record(stat *UnitStat) error {
	r.mutex.Lock()
	_, err := r.writer.WriteString(strx.JsonMarshal(stat) + "\n")
	r.mutex.Unlock()
	if err != nil {
		return errors.Wrap(err, "writer.WriteString failed")
	}
	return nil
}
