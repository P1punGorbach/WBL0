package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"zadaniel0/internal/cache"
	"zadaniel0/internal/config"
	"zadaniel0/internal/consumer"
	"zadaniel0/internal/http"
	"zadaniel0/internal/model"
	"zadaniel0/internal/storage"
)

type App struct {
	consumer *consumer.Consumer
	pg       *storage.Pg
	cc       *cache.Cache
	hh       *http.Http
}

func NewApp(ctx context.Context) (*App, error) {
	cfgfile, err := os.ReadFile("config.json")
	if err != nil {
		return nil, fmt.Errorf("Fail to readfile: %w", err)
	}
	var cfg config.Config
	err = json.Unmarshal(cfgfile, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Fail to connect: %w", err)
	}
	ch := make(chan model.Order)
	conn, err := consumer.NewConsumer(ctx, cfg.Kafka, ch)
	if err != nil {
		return nil, fmt.Errorf("Fail to create app: %w", err)
	}
	pg, err := storage.NewPg(ctx, cfg.Pg, ch)
	if err != nil {
		return nil, fmt.Errorf("Fail to create app: %w", err)
	}
	cc := cache.NewCache(pg)

	hh := http.NewHttp(cc)
	return &App{
		consumer: conn,
		pg:       pg,
		cc:       cc,
		hh:       hh,
	}, nil

}
func (A *App) Run(ctx context.Context) error {

	fmt.Println("Setup cache")
	err := A.cc.Setup(ctx)
	if err != nil {
		return fmt.Errorf("Fail to setup cache: %w", err)
	}

	fmt.Println("Setup http")
	A.hh.Setup()

	fmt.Println("Run Consumer")
	err = A.consumer.Run(ctx)

	if err != nil {
		return fmt.Errorf("Fail to run: %w", err)
	}
	fmt.Println("Run Pg")
	err = A.pg.Run(ctx)
	if err != nil {
		return fmt.Errorf("Fail to create run: %w", err)
	}

	fmt.Println("Run Http")
	err = A.hh.Run()
	if err != nil {
		return fmt.Errorf("Fail to run: %w", err)
	}

	return nil
}
