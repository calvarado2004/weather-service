package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/calvarado2004/weather-service/models"
	amqp "github.com/rabbitmq/amqp091-go"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	ctx := context.Background()

	type sendToQueueFunc func(service RabbitMQService, data interface{}, ctx context.Context) error

	var mySendToQueue sendToQueueFunc = sendToQueue

	// You can extract the API key from the environment variable here
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Fatalf("WEATHER_API_KEY environment variable not set")
	}

	// Set the real Weather API URLs
	currentWeatherURL := "https://api.weatherapi.com/v1/current.json"
	forecastWeatherURL := "https://api.weatherapi.com/v1/forecast.json"

	// Get weather data, loop every 15 minutes
	for {
		err := retrieveWeather(ctx, apiKey, currentWeatherURL, forecastWeatherURL, mySendToQueue)
		if err != nil {
			log.Printf("Error retrieving weather data: %s\n", err)
		}
		time.Sleep(15 * time.Minute)
	}

}

// retrieveWeather retrieves the weather data from the Weather API and sends it to the RabbitMQ queue
func retrieveWeather(ctx context.Context, apiKey, currentWeatherURL, forecastWeatherURL string, sendFunc func(service RabbitMQService, data interface{}, context context.Context) error) error {
	City := "Atlanta"

	weatherData, err := getWeatherData(apiKey, City, currentWeatherURL)
	if err != nil {
		return fmt.Errorf("error getting weather data: %w", err)
	}

	fmt.Println("Location:", weatherData.Location.Name, weatherData.Location.Region, weatherData.Location.Country)
	fmt.Println("Temperature:", weatherData.Current.TemperatureC, "°C")
	fmt.Println("Weather Condition:", weatherData.Current.Condition.Text)

	forecastResponse, err := getWeatherAlerts(apiKey, City, forecastWeatherURL)
	if err != nil {
		return fmt.Errorf("error getting weather alerts: %w", err)
	}

	if len(forecastResponse.Alerts.Alert) > 0 {
		fmt.Println("Weather Alert:", forecastResponse.Alerts.Alert[0])
	} else {
		fmt.Println("Weather Alert: No weather alerts")
	}

	// Calculate cellWeight
	cellWeight := assignWeight(weatherData.Current.Condition.Text)

	rabbitMessage := models.RabbitMQMessage{
		WeatherResponse:  *weatherData,
		ForecastResponse: *forecastResponse,
		CellWeight:       cellWeight,
	}

	rabbitService := &RealRabbitMQService{}
	err = sendFunc(rabbitService, rabbitMessage, ctx)
	if err != nil {
		return fmt.Errorf("error sending message to queue: %w", err)
	}

	log.Printf("Sent weather message to queue: %v\n", rabbitMessage)

	return nil
}

// assignWeight assigns a weight to the weather condition
func assignWeight(conditionText string) int {
	// source https://www.weatherapi.com/docs/weather_conditions.json

	switch conditionText {
	case "Sunny", "Clear", "Partly cloudy", "Cloudy", "Overcast", "Mist", "Fog", "Light Drizzle":
		return 1
	case "Patchy rain possible", "Patchy light drizzle", "Light rain", "Patchy light rain", "Light rain shower",
		"Light sleet showers":
		return 2
	case "Heavy freezing drizzle", "Heavy rain", "Moderate rain at times", "Moderate rain", "Heavy rain at times",
		"Heavy rain shower", "Torrential rain shower", "Light showers of ice pellets", "Freezing fog":
		return 3
	case "Thundery outbreaks possible", "Moderate or heavy snow showers", "Moderate or heavy snow with thunder",
		"Patchy moderate snow", "Moderate snow", "Patchy heavy snow", "Heavy snow", "Light snow", "Patchy light snow",
		"Light snow showers", "Blowing snow", "Blizzard", "Ice pellets", "Moderate or heavy showers of ice pellets":
		return 4
	default:
		return 0
	}
}

// getWeatherData makes a call to the Weather API and returns the response
func getWeatherData(apiKey string, City string, baseURL string) (*models.WeatherResponse, error) {

	url := fmt.Sprintf("%s?key=%s&q=%s&aqi=no", baseURL, apiKey, City)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making request: %s\n", err)
		os.Exit(1)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %s\n", err)
			os.Exit(1)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	var weatherData models.WeatherResponse
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %s", err)
	}

	return &weatherData, nil

}

// getWeatherAlerts makes a call to the Weather API and returns the response
func getWeatherAlerts(apiKey string, City string, baseURL string) (*models.ForecastResponse, error) {

	url := fmt.Sprintf("%s?key=%s&q=%s&days=1&aqi=no&alerts=yes", baseURL, apiKey, City)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making request: %s\n", err)
		os.Exit(1)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %s\n", err)
			os.Exit(1)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API returned non-200 status code: %d\n", resp.StatusCode)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s\n", err)
		return nil, err
	}

	var forecastResponse models.ForecastResponse
	err = json.Unmarshal(body, &forecastResponse)
	if err != nil {
		return nil, err
	}

	return &forecastResponse, nil

}

// sendToQueue sends the data to RabbitMQ
func sendToQueue(service RabbitMQService, data interface{}, context context.Context) error {

	// Set up connection to RabbitMQ, use environment variables or set defaults
	rabbitMQHost := os.Getenv("RABBITMQ_HOST")
	rabbitMQUser := os.Getenv("RABBITMQ_USER")
	rabbitMQPassword := os.Getenv("RABBITMQ_PASSWORD")
	rabbitMQPort := os.Getenv("RABBITMQ_PORT")

	if rabbitMQHost == "" {
		rabbitMQHost = "localhost"
	}

	if rabbitMQUser == "" {
		rabbitMQUser = "guest"
	}

	if rabbitMQPassword == "" {
		rabbitMQPassword = "guest"
	}

	if rabbitMQPort == "" {
		rabbitMQPort = "5672"
	}

	rabbitMQURL := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitMQUser, rabbitMQPassword, rabbitMQHost, rabbitMQPort)

	conn, err := service.Dial(rabbitMQURL)
	if err != nil {
		return err
	}
	defer func(conn AMQPConnection) {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Error closing connection: %s\n", err)
			os.Exit(1)
		}
	}(conn)

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	// Adjust this defer statement to accept the Channel interface, not *amqp.Channel
	defer func(ch Channel) {
		err := ch.Close()
		if err != nil {
			fmt.Printf("Error closing channel: %s\n", err)
			os.Exit(1)
		}
	}(ch)

	q, err := ch.QueueDeclare(
		"weather_queue", // Name of the queue
		false,           // Durable
		false,           // Delete when unused
		false,           // Exclusive
		false,           // No-wait
		nil,             // Arguments
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(
		context,
		"",     // Exchange
		q.Name, // Routing key
		false,  // Mandatory
		false,  // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	return err

}
