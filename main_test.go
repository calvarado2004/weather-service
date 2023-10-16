package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAssignWeight(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"Sunny", 1},
		{"Patchy rain possible", 2},
		{"Heavy freezing drizzle", 3},
		{"Thundery outbreaks possible", 4},
		{"Unknown Condition", 0},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			output := assignWeight(test.input)
			assert.Equal(t, test.expected, output)
		})
	}
}

func TestGetWeatherData(t *testing.T) {

	weatherSampleResponse := `{"location":{"name":"Roswell","region":"Georgia","country":"United States of America","lat":34.02,"lon":-84.36,"tz_id":"America/New_York","localtime_epoch":1697457029,"localtime":"2023-10-16 7:50"},"current":{"last_updated_epoch":1697456700,"last_updated":"2023-10-16 07:45","temp_c":7.2,"temp_f":45.0,"is_day":1,"condition":{"text":"Sunny","icon":"//cdn.weatherapi.com/weather/64x64/day/113.png","code":1000},"wind_mph":8.1,"wind_kph":13.0,"wind_degree":330,"wind_dir":"NNW","pressure_mb":1013.0,"pressure_in":29.91,"precip_mm":0.0,"precip_in":0.0,"humidity":86,"cloud":0,"feelslike_c":5.1,"feelslike_f":41.1,"vis_km":16.0,"vis_miles":9.0,"uv":1.0,"gust_mph":11.8,"gust_kph":18.9}}`

	// Mock the HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := rw.Write([]byte(weatherSampleResponse))
		if err != nil {
			t.Errorf("Error writing response: %s", err)
			return
		}
	}))
	defer server.Close()

	apiKey := "testApiKey"
	City := "Atlanta"

	data, err := getWeatherData(apiKey, City, server.URL)
	if err != nil {
		t.Fatalf("Error fetching weather data: %s", err)
		return
	}

	assert.Nil(t, err)

	assert.Equal(t, "Roswell", data.Location.Name)
	assert.Equal(t, "Georgia", data.Location.Region)
	assert.Equal(t, "United States of America", data.Location.Country)
	assert.Equal(t, 34.02, data.Location.Latitude)
	assert.Equal(t, -84.36, data.Location.Longitude)
	assert.Equal(t, "America/New_York", data.Location.TimezoneID)
	assert.Equal(t, int64(1697457029), data.Location.LocaltimeEpoch)
	assert.Equal(t, "2023-10-16 7:50", data.Location.Localtime)
	assert.Equal(t, int64(1697456700), data.Current.LastUpdatedEpoch)
	assert.Equal(t, "2023-10-16 07:45", data.Current.LastUpdated)
	assert.Equal(t, 7.2, data.Current.TemperatureC)
	assert.Equal(t, 45.0, data.Current.TemperatureF)
	assert.Equal(t, 1, data.Current.IsDay)
	assert.Equal(t, "Sunny", data.Current.Condition.Text)
	assert.Equal(t, "//cdn.weatherapi.com/weather/64x64/day/113.png", data.Current.Condition.Icon)
	assert.Equal(t, 1000, data.Current.Condition.Code)
	assert.Equal(t, 8.1, data.Current.WindMPH)
	assert.Equal(t, 13.0, data.Current.WindKPH)
	assert.Equal(t, 330, data.Current.WindDegree)
	assert.Equal(t, "NNW", data.Current.WindDir)
	assert.Equal(t, 1013.0, data.Current.PressureMB)
	assert.Equal(t, 29.91, data.Current.PressureIN)
	assert.Equal(t, 0.0, data.Current.PrecipMM)
	assert.Equal(t, 0.0, data.Current.PrecipIN)
	assert.Equal(t, 86, data.Current.Humidity)
	assert.Equal(t, 0, data.Current.Cloud)
	assert.Equal(t, 5.1, data.Current.FeelsLikeC)
	assert.Equal(t, 41.1, data.Current.FeelsLikeF)
	assert.Equal(t, 16.0, data.Current.VisibilityKM)
	assert.Equal(t, 9.0, data.Current.VisibilityMiles)
	assert.Equal(t, 1.0, data.Current.UV)
	assert.Equal(t, 11.8, data.Current.GustMPH)
	assert.Equal(t, 18.9, data.Current.GustKPH)

	return

}

func TestGetWeatherAlerts(t *testing.T) {

	forecastSampleResponse := `{"location":{"name":"Roswell","region":"Georgia","country":"United States of America","lat":34.02,"lon":-84.36,"tz_id":"America/New_York"},"alerts":{"alert":[{ "headline":"Flood Warning issued January 05 at 9:47PM EST until January 07 at 6:15AM EST by NWS", "msgtype":"Alert", "severity":"Moderate","urgency":"Expected","areas":"Roswell; Alpharetta","category":"Met","certainty":"Likely","event":"Flood Warning","note":"Alert for Roswell; Alpharetta (Georgia) Issued by the National Weather Service","effective":"2021-01-05T21:47:00-05:00","expires":"2021-01-07T06:15:00-05:00","desc":"...The Flood Warning continues"}]}}`

	// Mock the HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := rw.Write([]byte(forecastSampleResponse))
		if err != nil {
			t.Errorf("Error writing response: %s", err)
			return
		}
	}))
	defer server.Close()

	apiKey := "testApiKey"
	City := "Atlanta"

	data, err := getWeatherAlerts(apiKey, City, server.URL)
	if err != nil {
		t.Fatalf("Error fetching weather data: %s", err)
		return
	}
	assert.Nil(t, err)

	assert.Equal(t, "Flood Warning issued January 05 at 9:47PM EST until January 07 at 6:15AM EST by NWS", data.Alerts.Alert[0].Headline)
	assert.Equal(t, "Flood Warning", data.Alerts.Alert[0].Event)
	assert.Equal(t, "Moderate", data.Alerts.Alert[0].Severity)
	assert.Equal(t, "Likely", data.Alerts.Alert[0].Certainty)
	assert.Equal(t, "Alert for Roswell; Alpharetta (Georgia) Issued by the National Weather Service", data.Alerts.Alert[0].Note)
	assert.Equal(t, "2021-01-05T21:47:00-05:00", data.Alerts.Alert[0].Effective)
	assert.Equal(t, "2021-01-07T06:15:00-05:00", data.Alerts.Alert[0].Expires)
	assert.Equal(t, "...The Flood Warning continues", data.Alerts.Alert[0].Desc)

	return

}

type MockRabbitMQService struct {
	failOnPublish bool
	mock.Mock
}

func (m *MockRabbitMQService) Dial(url string) (AMQPConnection, error) {
	return &MockAMQPConnection{}, nil
}

func (m *MockRabbitMQService) Channel(conn AMQPConnection) (Channel, error) {
	return &MockChannel{failOnPublish: m.failOnPublish}, nil
}

func (m *MockRabbitMQService) QueueDeclare(ch Channel, name string) (amqp.Queue, error) {
	return amqp.Queue{Name: "mockQueue"}, nil
}

func (m *MockRabbitMQService) Publish(ch Channel, queueName string, body []byte) error {
	if m.failOnPublish {
		return errors.New("mock publish error")
	}
	return nil
}

type MockAMQPConnection struct{}

func (m *MockAMQPConnection) Close() error {
	return nil
}

func (m *MockAMQPConnection) Channel() (Channel, error) {
	return &MockChannel{}, nil
}

type MockChannel struct {
	failOnPublish bool
}

func (m *MockChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return amqp.Queue{Name: "mockQueue"}, nil
}

func (m *MockChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	if m.failOnPublish {
		return errors.New("mock publish error")
	}
	return nil
}

func (m *MockChannel) Close() error {
	return nil
}

func TestRetrieveWeather(t *testing.T) {

	ctx := context.Background()

	weatherSampleResponse := `{"location":{"name":"Roswell","region":"Georgia","country":"United States of America","lat":34.02,"lon":-84.36,"tz_id":"America/New_York","localtime_epoch":1697457029,"localtime":"2023-10-16 7:50"},"current":{"last_updated_epoch":1697456700,"last_updated":"2023-10-16 07:45","temp_c":7.2,"temp_f":45.0,"is_day":1,"condition":{"text":"Sunny","icon":"//cdn.weatherapi.com/weather/64x64/day/113.png","code":1000},"wind_mph":8.1,"wind_kph":13.0,"wind_degree":330,"wind_dir":"NNW","pressure_mb":1013.0,"pressure_in":29.91,"precip_mm":0.0,"precip_in":0.0,"humidity":86,"cloud":0,"feelslike_c":5.1,"feelslike_f":41.1,"vis_km":16.0,"vis_miles":9.0,"uv":1.0,"gust_mph":11.8,"gust_kph":18.9}}`

	forecastSampleResponse := `{"location":{"name":"Roswell","region":"Georgia","country":"United States of America","lat":34.02,"lon":-84.36,"tz_id":"America/New_York"},"alerts":{"alert":[{ "headline":"Flood Warning issued January 05 at 9:47PM EST until January 07 at 6:15AM EST by NWS", "msgtype":"Alert", "severity":"Moderate","urgency":"Expected","areas":"Roswell; Alpharetta","category":"Met","certainty":"Likely","event":"Flood Warning","note":"Alert for Roswell; Alpharetta (Georgia) Issued by the National Weather Service","effective":"2021-01-05T21:47:00-05:00","expires":"2021-01-07T06:15:00-05:00","desc":"...The Flood Warning continues"}]}}`

	// Mock the Weather API's response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/current.json":
			_, err := w.Write([]byte(weatherSampleResponse))
			if err != nil {
				t.Errorf("Error writing response: %s", err)
				return
			}
		case "/v1/forecast.json":
			_, err := w.Write([]byte(forecastSampleResponse))
			if err != nil {
				t.Errorf("Error writing response: %s", err)
				return
			}
		}
	}))
	defer server.Close()

	sendToQueueFunc := func(service RabbitMQService, data interface{}, context context.Context) error {

		service = &MockRabbitMQService{}
		// Ensure the provided service is the mock service (optional but can be a useful check)
		_, isMock := service.(*MockRabbitMQService)
		if !isMock {
			return fmt.Errorf("expected MockRabbitMQService, got %T", service)
		}

		body, err := json.Marshal(data)
		if err != nil {
			return err
		}

		t.Logf("Sending to queue: %s", string(body))

		return nil
	}

	ctx = context.Background()
	apiKey := "testApiKey"
	err := retrieveWeather(ctx, apiKey, server.URL+"/v1/current.json", server.URL+"/v1/forecast.json", sendToQueueFunc)

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestSendToQueue(t *testing.T) {
	mockService := new(MockRabbitMQService)
	mockConnection := new(MockAMQPConnection)
	mockChannel := new(MockChannel)

	mockService.On("Dial", mock.Anything).Return(mockConnection, nil)
	mockService.On("Channel", mockConnection).Return(mockChannel, nil)
	mockService.On("QueueDeclare", mockChannel, mock.Anything).Return(amqp.Queue{Name: "mockQueue"}, nil)

	// Test success
	mockService.failOnPublish = false
	err := sendToQueue(mockService, "{test message}", context.Background())
	assert.Nil(t, err)

}
