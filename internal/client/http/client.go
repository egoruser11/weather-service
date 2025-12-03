package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type client struct {
	httpClient http.Clien
}

type Response struct {
	Name      string `json:"name"`
	Country   string `json:"country"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

func NewClient(httpClient http.Client) *client {
	return &client{
		httpClient: httpClient,
	}
}

func (c *client) GetCoords(city string) (lat, lng float64, err error) {
	res, err := c.httpClient.Get(fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=ru&format=json", city))
	if err != nil {
		return 0, 0, err
	}

	if res.StatusCode != http.StatusOK {
		return 0, 0, errors.New(res.Status)
	}

	var geoResp struct {
		Results Response `json:"results"`
	}
	json.NewDecoder(res.Body).Decode(&geoResp)
	latitude, err := strconv.ParseFloat(geoResp.Results.Latitude, 64)
	if err != nil {
		return 0, 0, err
	}
	longitude, err := strconv.ParseFloat(geoResp.Results.Longitude, 64)
	if err != nil {
		return 0, 0, err
	}
	return latitude, longitude, nil
}
