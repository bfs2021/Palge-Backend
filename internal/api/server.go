package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *Application) Run() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.cfg.Port),
		Handler:           app.routes(),
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       time.Minute,
		ErrorLog:          slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	shutdownErr := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			shutdownErr <- err
		}

		if app.cfg.Limiter.Enabled {
			app.rateLimiter.StopJanitor()
		}

		app.logger.Info("completing background tasks", "addr", srv.Addr)

		app.wg.Wait()
		shutdownErr <- nil
	}()

	app.logger.Info(
		"starting server",
		"addr", srv.Addr,
		"env", app.cfg.Env,
	)

	if app.cfg.Limiter.Enabled {
		app.rateLimiter.StartJanitor(
			time.Duration(app.cfg.Limiter.JanitorInterval),
			time.Duration(app.cfg.Limiter.ClientExpiry),
		)
	}

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErr
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)
	return nil
}
