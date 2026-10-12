package api

import (
	"database/sql"
	"log/slog"
	"os"
	"sync"

	"github.com/nongpal/Palge-Backend/internal/account"
	"github.com/nongpal/Palge-Backend/internal/audit"
	"github.com/nongpal/Palge-Backend/internal/configs"
	"github.com/nongpal/Palge-Backend/internal/mailer"
	"github.com/nongpal/Palge-Backend/internal/permission"
	"github.com/nongpal/Palge-Backend/internal/token"
	"github.com/nongpal/Palge-Backend/internal/user"
)

const version = "0.5.0"

type Application struct {
	cfg         configs.Config
	rateLimiter *RateLimiter
	logger      *slog.Logger
	db          *sql.DB
	accounts    account.Model
	users       user.Model
	tokens      token.Model
	permissions permission.Model
	auditLogs   audit.Model
	mailer      *mailer.Mailer
	wg          sync.WaitGroup
}

func NewApplication(cfg configs.Config) (*Application, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	mailer, err := mailer.New(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Sender)
	if err != nil {
		return nil, err
	}

	app := &Application{
		cfg:         cfg,
		rateLimiter: NewRateLimiter(cfg.Limiter.RPS, cfg.Limiter.Burst),
		logger:      logger,
		db: db,
		accounts: account.Model{},
		users: user.Model{},
		tokens: token.Model{},
		permissions: permission.Model{},
		auditLogs: audit.Model{},
		mailer:      mailer,
	}
	return app, nil
}

func (app *Application) Close() error {
	return app.accounts.DB.Close()
}
