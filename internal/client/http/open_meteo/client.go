package open_meteo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Current struct {
		Time          string  `json:"time"`
		Temperature2m float64 `json:"temperature_2m"`
	}
}
type client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *client {
	return &client{
		httpClient: httpClient,
	}
}

func (c *client) GetTemperature(lat, lon float64) (Response, error) {
	res, err := c.httpClient.Get(
		fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m", lat, lon),
	)
	if err != nil {
		return Response{}, err
	}
	defer res.Body.Close()

	var response Response

	err = json.NewDecoder(res.Body).Decode(&response)

	if err != nil {
		return Response{}, err
	}
	return response, nil
}
