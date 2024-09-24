package kql

import "time"

type CustomMetricBody struct {
	Time string `json:"time"`
	Data struct {
		BaseData struct {
			Metric    string               `json:"metric"`
			Namespace string               `json:"namespace"`
			DimNames  []string             `json:"dimNames"`
			Series    []customMetricValues `json:"series"`
		} `json:"baseData"`
	} `json:"data"`
}

type customMetricValues struct {
	DimValues []string `json:"dimValues"`
	Min       int      `json:"min"`
	Max       int      `json:"max"`
	Sum       int      `json:"sum"`
	Count     int      `json:"count"`
}

func NewCustomMetricsBody(metricName string, metricValue float64) CustomMetricBody {
	body := CustomMetricBody{}

	body.Time = time.Now().Format(time.RFC3339)
	body.Data.BaseData.Metric = metricName
	body.Data.BaseData.Namespace = "CustomMetrics"
	body.Data.BaseData.DimNames = []string{metricName}
	body.Data.BaseData.Series = []customMetricValues{
		{
			DimValues: []string{metricName},
			Min:       int(metricValue),
			Max:       int(metricValue),
			Sum:       int(metricValue),
			Count:     1,
		},
	}
	return body
}