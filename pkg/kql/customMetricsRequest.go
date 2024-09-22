package kql

type customMetricBody struct {
	Time string `json:"time"`
	Data struct {
		BaseData struct {
			Metric    string   `json:"metric"`
			Namespace string   `json:"namespace"`
			DimNames  []string `json:"dimNames"`
			Series    []struct {
				DimValues []string `json:"dimValues"`
				Min       int      `json:"min"`
				Max       int      `json:"max"`
				Sum       int      `json:"sum"`
				Count     int      `json:"count"`
			} `json:"series"`
		} `json:"baseData"`
	} `json:"data"`
}