package monitor

import (
	"fmt"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/spf13/cast"

	"github.com/hatlonely/benv2/internal/recorder"
)

type ACMMonitorOptions struct {
	AccessKeyId     string
	AccessKeySecret string
	RegionId        string
	Endpoint        string
}

func NewACMMonitorWithOptions(options *ACMMonitorOptions) (*ACMMonitor, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(options.AccessKeyId),
		AccessKeySecret: tea.String(options.AccessKeySecret),
	}
	if options.RegionId != "" {
		config.Endpoint = tea.String(fmt.Sprintf("cms-export.%s.aliyuncs.com", options.RegionId))
	}
	if options.Endpoint != "" {
		config.Endpoint = tea.String(options.Endpoint)
	}

	cli, err := openapi.NewClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "openapi.NewClient failed")
	}

	return &ACMMonitor{
		options: options,
		client:  cli,
	}, nil
}

type ACMMonitor struct {
	options *ACMMonitorOptions
	client  *openapi.Client
}

func (m *ACMMonitor) Collect(startTime time.Time, endTime time.Time) (map[string][]*recorder.Measurement, error) {
	measurementMap := map[string][]*recorder.Measurement{}

	return measurementMap, nil
}

type Matcher struct {
	Label    string
	Value    string
	Operator string
}

type MeasurementReq struct {
	Namespace string
	Metric    string
	Period    time.Duration
	StartTime time.Time
	EndTime   time.Time
	Matchers  []Matcher
}

func (m *ACMMonitor) CollectOnce(req *MeasurementReq) ([]*recorder.Measurement, error) {
	cursorParams := &openapi.Params{
		Action:      tea.String("Cursor"),
		Version:     tea.String("2021-11-01"),
		Protocol:    tea.String("HTTPS"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("RPC"),
		Pathname:    tea.String("/"),
		ReqBodyType: tea.String("formData"),
		BodyType:    tea.String("json"),
	}

	cursorBody := map[string]interface{}{}
	cursorBody["Namespace"] = tea.String(req.Namespace)
	cursorBody["Metric"] = tea.String(req.Metric)
	cursorBody["Period"] = tea.Int(int(req.Period / time.Second))
	cursorBody["StartTime"] = tea.Int(int(req.StartTime.Unix() * 1000))
	cursorBody["EndTime"] = tea.Int(int(req.EndTime.Unix() * 1000))
	var matchers []map[string]*string
	for _, matcher := range req.Matchers {
		matchers = append(matchers, map[string]*string{
			"Label":    tea.String(matcher.Label),
			"Operator": tea.String(matcher.Operator),
			"Value":    tea.String(matcher.Value),
		})
	}
	cursorBody["Matchers"] = openapiutil.ArrayToStringWithSpecifiedStyle(matchers, tea.String("Matchers"), tea.String("json"))

	cursorResV, err := m.client.CallApi(cursorParams, &openapi.OpenApiRequest{
		Body: cursorBody,
	}, &util.RuntimeOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "m.client.CallApi failed")
	}

	var cursorRes struct {
		Success bool
		Code    int
		Data    struct {
			Cursor string
		}
		Message   string
		RequestId string
	}

	buf, _ := jsoniter.Marshal(cursorResV["body"])
	_ = jsoniter.Unmarshal(buf, &cursorRes)

	// batchGet
	batchGetParams := &openapi.Params{
		Action:      tea.String("BatchGet"),
		Version:     tea.String("2021-11-01"),
		Protocol:    tea.String("HTTPS"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("RPC"),
		Pathname:    tea.String("/"),
		ReqBodyType: tea.String("formData"),
		BodyType:    tea.String("json"),
	}
	batchGetBody := map[string]interface{}{}
	batchGetBody["Namespace"] = tea.String(req.Namespace)
	batchGetBody["Metric"] = tea.String(req.Metric)
	batchGetBody["Cursor"] = tea.String(cursorRes.Data.Cursor)
	batchGetBody["Length"] = tea.Int32(100000)

	batchGetResV, err := m.client.CallApi(batchGetParams, &openapi.OpenApiRequest{
		Body: batchGetBody,
	}, &util.RuntimeOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "m.client.CallApi failed")
	}

	var batchGetRes struct {
		Success bool
		Code    int
		Data    struct {
			Cursor  string
			Length  int
			Records []struct {
				Namespace     string
				Metric        string
				MeasureLabels []string
				MeasureValues []string
				Labels        []string
				LabelValues   []string
				Tags          []string
				TagValues     []string
				Timestamp     int
				Period        int
			}
		}
		Message   string
		RequestId string
	}

	buf, _ = jsoniter.Marshal(batchGetResV["body"])
	_ = jsoniter.Unmarshal(buf, &batchGetRes)
	//fmt.Println(batchGetResV["body"])
	//fmt.Println(batchGetRes.Data.Records)

	var measurements []*recorder.Measurement
	for _, record := range batchGetRes.Data.Records {
		kvs := map[string]float64{}
		for idx, key := range record.MeasureLabels {
			kvs[key] = cast.ToFloat64(record.MeasureValues[idx])
		}

		val, ok := kvs["Average"]
		if !ok {
			val = cast.ToFloat64(record.MeasureValues[0])
		}

		measurements = append(measurements, &recorder.Measurement{
			Time:  time.Unix(int64(record.Timestamp/1000), 0),
			Value: val,
		})
	}

	return measurements, nil
}

func MakeParams() *openapi.Params {
	return &openapi.Params{
		Action:      tea.String("Cursor"),
		Version:     tea.String("2021-11-01"),
		Protocol:    tea.String("HTTPS"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("RPC"),
		Pathname:    tea.String("/"),
		ReqBodyType: tea.String("formData"),
		BodyType:    tea.String("json"),
	}
}
