package driver

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type HttpDriverOptions struct {
	DialTimeout         time.Duration `dft:"3s"`
	Timeout             time.Duration `dft:"6s"`
	MaxIdleConnsPerHost int           `dft:"2"`
}

func NewHttpDriverWithOptions(options *HttpDriverOptions) (*HttpDriver, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, options.DialTimeout)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			MaxIdleConnsPerHost: options.MaxIdleConnsPerHost,
		},
		Timeout: options.Timeout,
	}

	return &HttpDriver{
		client: client,
	}, nil
}

type HttpDriver struct {
	client *http.Client
}

type HttpDoReq struct {
	Method     string
	URL        string
	Params     map[string]string
	Headers    map[string]string
	Data       string
	Json       interface{}
	Timeout    time.Duration
	JsonDecode bool
}

type HttpDoRes struct {
	Status  int
	Headers map[string]string
	Json    interface{}
	Text    string
}

func (d *HttpDriver) Do(req *HttpDoReq) (*HttpDoRes, error) {
	var buf []byte
	if req.Json != nil {
		var err error
		buf, err = jsoniter.Marshal(req.Json)
		if err != nil {
			return nil, errors.WithMessage(err, "jsoniter.Marshal failed")
		}
	}
	if req.Data != "" {
		buf = []byte(req.Data)
	}

	hreq, err := http.NewRequest(req.Method, req.URL, bytes.NewReader(buf))
	if err != nil {
		return nil, errors.WithMessage(err, "http.NewRequest failed")
	}

	for key, val := range req.Headers {
		hreq.Header.Set(key, val)
	}

	if req.Params != nil {
		q := hreq.URL.Query()
		for key, val := range req.Params {
			q.Add(key, val)
		}
		hreq.URL.RawQuery = q.Encode()
	}

	hres, err := d.client.Do(hreq)
	if err != nil {
		return nil, errors.Wrap(err, "client.Do failed")
	}
	defer hres.Body.Close()

	res := &HttpDoRes{
		Status: hres.StatusCode,
	}

	if hres.Header != nil {
		res.Headers = map[string]string{}
		for key := range hres.Header {
			res.Headers[key] = hres.Header.Get(key)
		}
	}

	buf, err = ioutil.ReadAll(hres.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil.ReadAll failed")
	}

	if req.JsonDecode {
		if err := jsoniter.Unmarshal(buf, &res.Json); err != nil {
			return nil, errors.Wrap(err, "jsoniter.Unmarshal failed")
		}
	} else {
		res.Text = string(buf)
	}

	return res, nil
}
