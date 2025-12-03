package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
	"log"
	"net/http"
	"sync"
	"time"
	"weather-service/internal/client/http/geocoding"
	"weather-service/internal/client/http/open_meteo"
)

const httpPort = ":3000"

func main() {

	r := chi.NewRouter()

	httpClient := &http.Client{}

	geocodingClient := geocoding.NewClient(httpClient)
	openMeteoClient := open_meteo.NewClient(httpClient)

	r.Use(middleware.Logger)
	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		res, err := geocodingClient.GetCoords(city)
		if err != nil {
			log.Println(err)
			return
		}

		resp, err := openMeteoClient.GetTemperature(res.Latitude, res.Longitude)
		if err != nil {
			log.Println(err)
			return
		}
		tempRes, err := json.Marshal(resp)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = w.Write(tempRes)
	})

	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}
	jobs, err := initJobs(s)
	if err != nil {
		panic(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		fmt.Printf("starting job: %v\n ", jobs[0].ID())
		s.Start()
	}()

	go func() {
		defer wg.Done()

		fmt.Println("starting serve")
		err := http.ListenAndServe(httpPort, r)
		if err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}
func initJobs(scheduler gocron.Scheduler) ([]gocron.Job, error) {
	j, err := scheduler.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				fmt.Println("hello world")
			},
		),
	)
	if err != nil {
		return nil, err
	}
	return []gocron.Job{j}, nil
}
