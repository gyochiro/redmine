package prome

import (
	"testing"
	"time"
)

func TestRemoteWrite(t *testing.T) {
	c, err := NewClient("http://localhost:9090/api/v1/write", 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	metrics := []MetricPoint{
		{Metric: "opcai10",
			TagsMap: map[string]string{"env": "testing", "op": "opcai"},
			Time:    time.Now().Add(-15 * time.Minute).Unix(),
			Value:   1},
		{Metric: "opcai9",
			TagsMap: map[string]string{"env": "testing", "op": "opcai"},
			Time:    time.Now().Add(-3 * time.Minute).Unix(),
			Value:   2},
		{Metric: "opcai5",
			TagsMap: map[string]string{"env": "testing", "op": "opcai"},
			Time:    time.Now().Unix(),
			Value:   30},
		{Metric: "opcai6",
			TagsMap: map[string]string{"env": "testing", "op": "opcai"},
			Time:    time.Now().Unix(),
			Value:   40},
	}
	err = c.RemoteWrite(metrics)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("end...")
}
