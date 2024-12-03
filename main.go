package main

import (
	"context"
	"zadaniel0/internal/app"
)

func main() {
	app, err := app.NewApp(context.Background())
	if err != nil {
		panic(err)
	}
	err = app.Run(context.Background())
	if err != nil {
		panic(err)
	}
}