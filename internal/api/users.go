package api

import (
	"net/http"
	"time"

	"github.com/nongpal/Palge-Backend/internal/apperr"
	"github.com/nongpal/Palge-Backend/internal/audit"
	"github.com/nongpal/Palge-Backend/internal/token"
	"github.com/nongpal/Palge-Backend/internal/user"
	"github.com/nongpal/Palge-Backend/internal/validator"
)

func (app *Application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.writeError(w, r, apperr.Wrap(apperr.CodeInvalidRequest, err))
		return
	}

	v := validator.New()

	usr := &user.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	if err := usr.Password.Set(input.Password); err != nil {
		app.writeError(w, r, err)
		return
	}

	user.ValidateUser(v, usr)
	if !v.Valid() {
		app.writeError(w, r, apperr.WithFields(apperr.CodeValidationFailed, v.Errors))
		return
	}

	if err := app.users.Insert(usr); err != nil {
		app.recordAuditLog(r, &audit.AuditLog{
			UserID:     nil,
			Action:     "register",
			Resource:   "user",
			ResourceID: nil,
			Result:     "failed",
		})

		app.writeError(w, r, err)
		return
	}

	token, err := app.tokens.New(usr.ID, 3*24*time.Hour, token.ScopeActivation)
	if err != nil {
		app.writeError(w, r, err)
		return
	}

	app.background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          usr.ID,
		}
		err = app.mailer.Send(usr.Email, "user_welcome.tmpl.html", data)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	app.recordAuditLog(r, &audit.AuditLog{
		UserID:     nil,
		Action:     "register",
		Resource:   "user",
		ResourceID: &usr.ID,
		Result:     "success",
	})

	if err := app.writeJSON(w, http.StatusAccepted, envelope{
		"user": usr,
	}, nil); err != nil {
		app.writeError(w, r, err)
	}
}

func (app *Application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.writeError(w, r, apperr.Wrap(apperr.CodeInvalidRequest, err))
		return
	}

	v := validator.New()

	if token.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.recordAuditLog(r, &audit.AuditLog{
			UserID:     nil,
			Action:     "activate",
			Resource:   "user",
			ResourceID: nil,
			Result:     "failed",
		})
		app.writeError(w, r, apperr.WithFields(apperr.CodeValidationFailed, v.Errors))
		return
	}

	user, err := app.users.GetForToken(token.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		app.recordAuditLog(r, &audit.AuditLog{
			UserID:     nil,
			Action:     "activate",
			Resource:   "user",
			ResourceID: nil,
			Result:     "failed",
		})

		app.writeError(w, r, err)
		return
	}

	user.Activated = true

	if err := app.users.Update(user); err != nil {
		app.writeError(w, r, err)
		return
	}

	if err := app.permissions.AddForUser(
		user.ID,
		"accounts:read",
		"accounts:create",
		"accounts:deposit",
		"accounts:withdraw",
		"accounts:transfer",
	); err != nil {
		app.writeError(w, r, err)
		return
	}

	if err := app.tokens.DeleteAllForUser(token.ScopeActivation, user.ID); err != nil {
		app.writeError(w, r, err)
		return
	}

	app.recordAuditLog(r, &audit.AuditLog{
		UserID:     &user.ID,
		Action:     "activate",
		Resource:   "user",
		ResourceID: &user.ID,
		Result:     "success",
	})

	if err := app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil); err != nil {
		app.writeError(w, r, err)
	}
}
