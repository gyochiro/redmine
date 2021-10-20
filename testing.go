package prome

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/prompb"
)

type RecoverableError struct {
	error
}

type HttpClient struct {
	url     *url.URL
	Client  *http.Client
	timeout time.Duration
}

var MetricNameRE = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)

type MetricPoint struct {
	Metric  string            `json:"metric"` // 指標名稱
	TagsMap map[string]string `json:"tags"`   // 資料標籤
	Time    int64             `json:"time"`   // 時間戳，單位是秒
	Value   float64           `json:"value"`  // 內部欄位，最終轉換之後的float64數值
}

func (c *HttpClient) remoteWritePost(req []byte) error {
	httpReq, err := http.NewRequest("POST", c.url.String(), bytes.NewReader(req))
	if err != nil {
		return err
	}
	httpReq.Header.Add("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("User-Agent", "opcai")
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	httpReq = httpReq.WithContext(ctx)

	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		var ht *nethttp.Tracer
		httpReq, ht = nethttp.TraceRequest(
			parentSpan.Tracer(),
			httpReq,
			nethttp.OperationName("Remote Store"),
			nethttp.ClientTrace(false),
		)
		defer ht.Finish()
	}

	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		// Errors from Client.Do are from (for example) network errors, so are
		// recoverable.
		return RecoverableError{err}
	}
	defer func() {
		io.Copy(ioutil.Discard, httpResp.Body)
		httpResp.Body.Close()
	}()

	if httpResp.StatusCode/100 != 2 {
		scanner := bufio.NewScanner(io.LimitReader(httpResp.Body, 512))
		line := ""
		if scanner.Scan() {
			line = scanner.Text()
		}
		err = errors.Errorf("server returned HTTP status %s: %s", httpResp.Status, line)
	}
	if httpResp.StatusCode/100 == 5 {
		return RecoverableError{err}
	}
	return err
}

func buildWriteRequest(samples []*prompb.TimeSeries) ([]byte, error) {

	req := &prompb.WriteRequest{
		Timeseries: samples,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	compressed := snappy.Encode(nil, data)
	return compressed, nil
}

type sample struct {
	labels labels.Labels
	t      int64
	v      float64
}

const (
	LABEL_NAME = "__name__"
)

func convertOne(item *MetricPoint) (*prompb.TimeSeries, error) {
	pt := prompb.TimeSeries{}
	pt.Samples = []prompb.Sample{{}}
	s := sample{}
	s.t = item.Time
	s.v = item.Value
	// name
	if !MetricNameRE.MatchString(item.Metric) {
		return &pt, errors.New("invalid metrics name")
	}
	nameLs := labels.Label{
		Name:  LABEL_NAME,
		Value: item.Metric,
	}
	s.labels = append(s.labels, nameLs)
	for k, v := range item.TagsMap {
		if model.LabelNameRE.MatchString(k) {
			ls := labels.Label{
				Name:  k,
				Value: v,
			}
			s.labels = append(s.labels, ls)
		}
	}

	pt.Labels = labelsToLabelsProto(s.labels, pt.Labels)
	// 時間賦值問題,使用毫秒時間戳
	tsMs := time.Unix(s.t, 0).UnixNano() / 1e6
	pt.Samples[0].Timestamp = tsMs
	pt.Samples[0].Value = s.v
	return &pt, nil
}

func labelsToLabelsProto(labels labels.Labels, buf []*prompb.Label) []*prompb.Label {
	result := buf[:0]
	if cap(buf) < len(labels) {
		result = make([]*prompb.Label, 0, len(labels))
	}
	for _, l := range labels {
		result = append(result, &prompb.Label{
			Name:  l.Name,
			Value: l.Value,
		})
	}
	return result
}

func (c *HttpClient) RemoteWrite(items []MetricPoint) (err error) {
	if len(items) == 0 {
		return
	}
	ts := make([]*prompb.TimeSeries, len(items))
	for i := range items {
		ts[i], err = convertOne(&items[i])
		if err != nil {
			return
		}
	}
	data, err := buildWriteRequest(ts)
	if err != nil {
		return
	}
	err = c.remoteWritePost(data)
	return
}

func NewClient(ur string, timeout time.Duration) (c *HttpClient, err error) {
	u, err := url.Parse(ur)
	if err != nil {
		return
	}
	c = &HttpClient{
		url:     u,
		Client:  &http.Client{},
		timeout: timeout,
	}
	return
}
