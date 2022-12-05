package recorder

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type FileAnalystOptions struct {
	FilePath string
}

func NewFileAnalystWithOptions(options *FileAnalystOptions) (*FileAnalyst, error) {
	return &FileAnalyst{
		options: options,
	}, nil
}

type FileAnalyst struct {
	options *FileAnalystOptions
}

func (fa *FileAnalyst) TimeRange() (time.Time, time.Time, error) {
	const kBufSize = 4096

	fp, err := os.Open(fa.options.FilePath)
	if err != nil {
		return time.Time{}, time.Time{}, errors.WithMessage(err, "os.Open failed")
	}
	defer fp.Close()

	var head bytes.Buffer
	// head -1
	{
		buf := make([]byte, kBufSize)
	headOut:
		for {
			n, err := fp.Read(buf)
			if err != nil && err != io.EOF {
				return time.Time{}, time.Time{}, errors.WithMessage(err, "os.Open failed")
			}
			if err == io.EOF {
				break
			}

			for i := 0; i < n; i++ {
				if buf[i] == '\n' {
					head.Write(buf[0:i])
					break headOut
				}
			}
			head.Write(buf[0:n])
		}
	}

	var tail bytes.Buffer
	// tail -1
	{
		var rtail bytes.Buffer
		buf := make([]byte, kBufSize)

		// 获取文件大小
		offset, err := fp.Seek(0, io.SeekEnd)
		if err != nil {
			return time.Time{}, time.Time{}, errors.Wrap(err, "fp.Seek failed")
		}

	tailOut:
		for i := 0; ; i++ {
			n := 0 // 将要读取的字节数
			if offset < kBufSize {
				offset = 0
				n = int(offset)
			} else {
				offset -= kBufSize
				n = kBufSize
			}

			_, err := fp.Seek(offset, io.SeekStart)
			if err != nil {
				return time.Time{}, time.Time{}, errors.Wrap(err, "fp.Seek failed")
			}

			n, err = fp.Read(buf)
			if err != nil && err != io.EOF {
				return time.Time{}, time.Time{}, errors.WithMessage(err, "os.Open failed")
			}

			for j := n - 1; j >= 0; j-- {
				if buf[j] == '\n' && (i != 0 || j != n-1) {
					break tailOut
				}
				rtail.WriteByte(buf[j])
			}

			if offset == 0 {
				break
			}
		}

		for i := rtail.Len() - 1; i >= 0; i-- {
			tail.WriteByte(rtail.Bytes()[i])
		}
	}

	var su UnitStat
	if err := jsoniter.Unmarshal(head.Bytes(), &su); err != nil {
		return time.Time{}, time.Time{}, errors.Wrap(err, "jsoniter.Unmarshal failed")
	}
	st, err := time.Parse(time.RFC3339Nano, su.Time)
	if err != nil {
		return time.Time{}, time.Time{}, errors.Wrap(err, "time.Parse failed")
	}

	var eu UnitStat
	if err := jsoniter.Unmarshal(tail.Bytes(), &eu); err != nil {
		return time.Time{}, time.Time{}, errors.Wrap(err, "jsoniter.Unmarshal failed")
	}
	et, err := time.Parse(time.RFC3339Nano, eu.Time)
	if err != nil {
		return time.Time{}, time.Time{}, errors.Wrap(err, "time.Parse failed")
	}

	return st, et, nil
}

type FileAnalystStatStream struct {
	reader *bufio.Reader
}

func (fa *FileAnalyst) Stat() (StatStream, error) {
	fp, err := os.Open(fa.options.FilePath)
	if err != nil {
		return nil, errors.Wrap(err, "os.Open failed")
	}
	reader := bufio.NewReader(fp)

	return &FileAnalystStatStream{
		reader: reader,
	}, nil
}

func (s *FileAnalystStatStream) Next() (*UnitStat, error) {
	buf, err := s.reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, errors.Wrapf(err, "reader.ReadBytes failed")
	}

	if err == io.EOF {
		return nil, nil
	}

	var v UnitStat
	if err := jsoniter.Unmarshal(buf, &v); err != nil {
		return nil, errors.Wrapf(err, "jsoniter.Unmarshal failed")
	}

	return &v, nil
}
