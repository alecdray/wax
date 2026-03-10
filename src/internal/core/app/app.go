package app

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNoValue = errors.New("no value")
)

type App struct {
	config Config
	claims *Claims
}

func NewApp(config Config) App {
	return App{config: config}
}

func (app App) Config() Config {
	return app.config
}

func (app *App) SetClaims(w http.ResponseWriter, claims *Claims) error {
	app.claims = claims
	err := app.claims.Save(app.config, w)
	if err != nil {
		return fmt.Errorf("failed to save JWT: %w", err)
	}
	return nil
}

func (app App) Claims() *Claims {
	return app.claims
}

func (app *App) DeleteClaims(w http.ResponseWriter) {
	app.claims.Delete(app.config, w)
	app.claims = nil
}
