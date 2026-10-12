package api

import (
	"fmt"
	"net/http"

	"github.com/nongpal/Palge-Backend/internal/account"
	"github.com/nongpal/Palge-Backend/internal/apperr"
	"github.com/nongpal/Palge-Backend/internal/audit"
	"github.com/nongpal/Palge-Backend/internal/validator"
)

func (app *Application) createAccountHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Owner          string `json:"owner"`
		InitialBalance int64  `json:"initial_balance"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.writeError(w, r, apperr.Wrap(apperr.CodeInvalidRequest, err))
		return
	}

	acc := &account.Account{
		Owner:   input.Owner,
		Balance: input.InitialBalance,
	}

	v := validator.New()

	account.ValidateAccount(v, acc)

	if !v.Valid() {
		app.writeError(w, r, apperr.WithFields(apperr.CodeValidationFailed, v.Errors))
		return
	}

	if err = app.accounts.Insert(r.Context(), acc); err != nil {
		app.writeError(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/accounts/%d", acc.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"account": acc}, headers)

	if err != nil {
		app.writeError(w, r, err)
	}
}

func (app *Application) listAccountHandler(w http.ResponseWriter, r *http.Request) {
	accounts, err := app.accounts.GetAll(r.Context())
	if err != nil {
		app.writeError(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{
		"account": accounts,
	}, nil)

	if err != nil {
		app.writeError(w, r, err)
	}
}

func (app *Application) showAccountHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.writeError(w, r, apperr.Wrap(apperr.CodeInvalidRequest, err))
		return
	}

	account, err := app.accounts.Get(r.Context(), id)
	if err != nil {
		app.writeError(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"account": account}, nil)
	if err != nil {
		app.writeError(w, r, err)
	}
}

func (app *Application) depositHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Amount int64 `json:"amount"`
	}

	id, err := app.readIDParam(r)
	if err != nil {
		app.writeError(w, r, apperr.Wrap(apperr.CodeInvalidRequest, err))
		return
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.writeError(w, r, apperr.Wrap(apperr.CodeInvalidRequest, err))
		return
	}

	v := validator.New()
	account.ValidateAmount(v, input.Amount)

	if !v.Valid() {
		app.writeError(w, r, apperr.WithFields(apperr.CodeValidationFailed, v.Errors))
		return
	}

	user := app.contextGetUser(r)

	account, err := app.accounts.Deposit(
		r.Context(),
		id,
		input.Amount,
	)

	if err != nil {
		app.recordAuditLog(r, &audit.AuditLog{
			UserID:     &user.ID,
			Action:     "deposit",
			Resource:   "account",
			ResourceID: &id,
			Result:     "failed",
		})

		app.writeError(w, r, err)
		return
	}

	app.recordAuditLog(r, &audit.AuditLog{
		UserID:     &user.ID,
		Action:     "deposit",
		Resource:   "account",
		ResourceID: &id,
		Result:     "success",
	})

	err = app.writeJSON(
		w,
		http.StatusOK,
		envelope{"account": account},
		nil,
	)
	if err != nil {
		app.writeError(w, r, err)
	}
}

func (app *Application) withdrawHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Amount int64 `json:"amount"`
	}

	id, err := app.readIDParam(r)
	if err != nil {
		app.writeError(w, r, apperr.Wrap(apperr.CodeInvalidRequest, err))
		return
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.writeError(w, r, apperr.Wrap(apperr.CodeInvalidRequest, err))
		return
	}

	v := validator.New()
	if account.ValidateAmount(v, input.Amount); !v.Valid() {
		app.writeError(w, r, apperr.WithFields(apperr.CodeValidationFailed, v.Errors))
		return
	}

	user := app.contextGetUser(r)

	account, err := app.accounts.Withdraw(r.Context(), id, input.Amount)
	if err != nil {
		app.recordAuditLog(r, &audit.AuditLog{
			UserID:     &user.ID,
			Action:     "withdraw",
			Resource:   "account",
			ResourceID: &id,
			Result:     "failed",
		})

		app.writeError(w, r, err)
		return
	}

	app.recordAuditLog(r, &audit.AuditLog{
		UserID:     &user.ID,
		Action:     "withdraw",
		Resource:   "account",
		ResourceID: &id,
		Result:     "success",
	})

	err = app.writeJSON(w, http.StatusOK, envelope{"account": account}, nil)
	if err != nil {
		app.writeError(w, r, err)
	}
}

func (app *Application) transferHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		From   int64 `json:"from"`
		To     int64 `json:"to"`
		Amount int64 `json:"amount"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.writeError(w, r, apperr.Wrap(apperr.CodeInvalidRequest, err))
		return
	}

	v := validator.New()
	account.ValidateAmount(v, input.Amount)
	v.Check(input.From > 0, "from", "must be a valid account ID")
	v.Check(input.To > 0, "to", "must be a valid account ID")
	v.Check(input.From != input.To, "to", "cannot transfer to the same account")

	if !v.Valid() {
		app.writeError(w, r, apperr.WithFields(apperr.CodeValidationFailed, v.Errors))
		return
	}

	user := app.contextGetUser(r)

	sender, receiver, err := app.accounts.Transfer(r.Context(), input.From, input.To, input.Amount)
	if err != nil {
		app.recordAuditLog(r, &audit.AuditLog{
			UserID:     &user.ID,
			Action:     "transfer",
			Resource:   "account",
			ResourceID: &input.From,
			Result:     "failed",
		})

		app.writeError(w, r, err)
		return
	}

	app.recordAuditLog(r, &audit.AuditLog{
		UserID:     &user.ID,
		Action:     "transfer",
		Resource:   "account",
		ResourceID: &input.From,
		Result:     "success",
	})

	err = app.writeJSON(w, http.StatusOK, envelope{
		"transfer": map[string]*account.Account{
			"from": sender,
			"to":   receiver,
		},
	}, nil)

	if err != nil {
		app.writeError(w, r, err)
	}
}
