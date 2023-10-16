package models

type WeatherResponse struct {
	Location struct {
		Name           string  `json:"name"`
		Region         string  `json:"region"`
		Country        string  `json:"country"`
		Latitude       float64 `json:"lat"`
		Longitude      float64 `json:"lon"`
		TimezoneID     string  `json:"tz_id"`
		LocaltimeEpoch int64   `json:"localtime_epoch"`
		Localtime      string  `json:"localtime"`
	} `json:"location"`
	Current struct {
		LastUpdatedEpoch int64   `json:"last_updated_epoch"`
		LastUpdated      string  `json:"last_updated"`
		TemperatureC     float64 `json:"temp_c"`
		TemperatureF     float64 `json:"temp_f"`
		IsDay            int     `json:"is_day"`
		Condition        struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
		WindMPH         float64 `json:"wind_mph"`
		WindKPH         float64 `json:"wind_kph"`
		WindDegree      int     `json:"wind_degree"`
		WindDir         string  `json:"wind_dir"`
		PressureMB      float64 `json:"pressure_mb"`
		PressureIN      float64 `json:"pressure_in"`
		PrecipMM        float64 `json:"precip_mm"`
		PrecipIN        float64 `json:"precip_in"`
		Humidity        int     `json:"humidity"`
		Cloud           int     `json:"cloud"`
		FeelsLikeC      float64 `json:"feelslike_c"`
		FeelsLikeF      float64 `json:"feelslike_f"`
		VisibilityKM    float64 `json:"vis_km"`
		VisibilityMiles float64 `json:"vis_miles"`
		UV              float64 `json:"uv"`
		GustMPH         float64 `json:"gust_mph"`
		GustKPH         float64 `json:"gust_kph"`
	} `json:"current"`
}

type Alert struct {
	Headline    string `json:"headline"`
	MsgType     string `json:"msgtype"`
	Severity    string `json:"severity"`
	Urgency     string `json:"urgency"`
	Areas       string `json:"areas"`
	Category    string `json:"category"`
	Certainty   string `json:"certainty"`
	Event       string `json:"event"`
	Note        string `json:"note"`
	Effective   string `json:"effective"`
	Expires     string `json:"expires"`
	Desc        string `json:"desc"`
	Instruction string `json:"instruction"`
}

type ForecastResponse struct {
	Alerts struct {
		Alert []Alert `json:"alert"`
	} `json:"alerts"`
}

type RabbitMQMessage struct {
	WeatherResponse  WeatherResponse
	ForecastResponse ForecastResponse
	CellWeight       int
}
