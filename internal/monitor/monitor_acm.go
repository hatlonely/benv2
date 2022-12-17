package monitor

import (
	"fmt"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/pkg/errors"

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
	params := &openapi.Params{
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

	body := map[string]interface{}{}
	body["Namespace"] = tea.String(req.Namespace)
	body["Metric"] = tea.String(req.Metric)
	body["Period"] = tea.Int(int(req.Period / time.Second))
	body["StartTime"] = tea.Int(int(req.StartTime.Unix() * 1000))
	body["EndTime"] = tea.Int(int(req.EndTime.Unix() * 1000))
	var matchers []map[string]*string
	for _, matcher := range req.Matchers {
		matchers = append(matchers, map[string]*string{
			"Label":    tea.String(matcher.Label),
			"Operator": tea.String(matcher.Operator),
			"Value":    tea.String(matcher.Value),
		})
	}
	body["Matchers"] = openapiutil.ArrayToStringWithSpecifiedStyle(matchers, tea.String("Matchers"), tea.String("json"))
	runtime := &util.RuntimeOptions{}
	request := &openapi.OpenApiRequest{
		Body: body,
	}

	res, err := m.client.CallApi(params, request, runtime)
	if err != nil {
		return nil, errors.Wrap(err, "m.client.CallApi failed")
	}

	fmt.Println(res)

	return nil, nil
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
