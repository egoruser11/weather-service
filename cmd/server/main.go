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

const (
	httpPort = ":3000"
	msc      = "moscow"
)

type Reading struct {
	Timestamp   time.Time
	Temperature float64
}
type Storage struct {
	data map[string][]Reading
	mu   sync.RWMutex
}

func main() {

	r := chi.NewRouter()
	storage := &Storage{
		data: make(map[string][]Reading),
	}

	r.Use(middleware.Logger)
	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		_ = city
		storage.mu.RLock()
		defer storage.mu.RUnlock()
		reading, ok := storage.data[city]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
			return
		}
		raw, err := json.Marshal(reading)
		if err != nil {
			fmt.Println(err)
			return
		}
		_, _ = w.Write(raw)

	})

	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}
	jobs, err := initJobs(s, storage)
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
func initJobs(scheduler gocron.Scheduler, storage *Storage) ([]gocron.Job, error) {

	httpClient := &http.Client{}
	geocodingClient := geocoding.NewClient(httpClient)
	openMeteoClient := open_meteo.NewClient(httpClient)
	j, err := scheduler.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				res, err := geocodingClient.GetCoords(msc)
				if err != nil {
					log.Println(err)
					return
				}

				resp, err := openMeteoClient.GetTemperature(res.Latitude, res.Longitude)
				if err != nil {
					log.Println(err)
					return
				}
				timestamp, err := time.Parse("2006-01-02T15:04", resp.Current.Time)
				if err != nil {
					log.Println(err)
					return
				}
				storage.mu.Lock()
				defer storage.mu.Unlock()
				storage.data[msc] = append(storage.data[msc], Reading{
					Timestamp:   timestamp,
					Temperature: resp.Current.Temperature2m,
				})
				fmt.Printf("updated temperature to %v\n ", resp.Current.Temperature2m)
				fmt.Printf("updated tumestamo to %s\n ", timestamp)
			},
		),
	)
	if err != nil {
		return nil, err
	}
	return []gocron.Job{j}, nil
}
