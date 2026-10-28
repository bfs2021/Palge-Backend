package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/nongpal/Palge-Backend/internal/account"
	"github.com/nongpal/Palge-Backend/internal/audit"
	"github.com/nongpal/Palge-Backend/internal/testutil"
	"github.com/nongpal/Palge-Backend/internal/user"
)

func TestAccountHandlersEnforceOwnership(t *testing.T) {
	db := testutil.NewTestDB(t)
	app := &Application{
		accounts:  account.Model{DB: db},
		auditLogs: audit.Model{DB: db},
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	seedUser := func(name string) int64 {
		var id int64
		query := "INSERT INTO users (name, email, password_hash, activated) VALUES ($1, $2, $3, true)" +
			" ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name RETURNING id"
		err := db.QueryRow(query, strings.ToUpper(name), name+"@palge.test", []byte("x")).Scan(&id)
		if err != nil {
			t.Fatalf("seed user: %v", err)
		}
		return id
	}

	// Accounts are created through the handler itself, so the user_id the
	// ownership checks read is the one Insert persisted.
	createAccount := func(userID int64, owner string) int64 {
		w := httptest.NewRecorder()
		body := fmt.Sprintf(`{"owner":%q,"initial_balance":100}`, owner)
		app.createAccountHandler(w, requestWithUser(userID, "", "create", body))
		if w.Code != http.StatusCreated {
			t.Fatalf("create account %s: status = %d", owner, w.Code)
		}
		var res struct {
			Account account.Account `json:"account"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("decode created account: %v", err)
		}
		return res.Account.ID
	}

	owner := seedUser("owner")
	intruder := seedUser("intruder")
	ownerAcc := createAccount(owner, "owner-acc")
	intruderAcc := createAccount(intruder, "intruder-acc")
	ownerID := strconv.FormatInt(ownerAcc, 10)
	transferOut := fmt.Sprintf(`{"from":%d,"to":%d,"amount":25}`, ownerAcc, intruderAcc)
	transferIn := fmt.Sprintf(`{"from":%d,"to":%d,"amount":25}`, intruderAcc, ownerAcc)

	tests := []struct {
		name   string
		status int
		handle func(http.ResponseWriter, *http.Request)
		req    *http.Request
		want   map[int64]int64
	}{
		{
			name:   "deposit into another user's account",
			status: http.StatusForbidden,
			handle: app.depositHandler,
			req:    requestWithUser(intruder, ownerID, "deposit", `{"amount":50}`),
			want:   map[int64]int64{ownerAcc: 100},
		},
		{
			name:   "withdraw from another user's account",
			status: http.StatusForbidden,
			handle: app.withdrawHandler,
			req:    requestWithUser(intruder, ownerID, "withdraw", `{"amount":50}`),
			want:   map[int64]int64{ownerAcc: 100},
		},
		{
			name:   "transfer debiting another user's account",
			status: http.StatusForbidden,
			handle: app.transferHandler,
			req:    requestWithUser(intruder, "", "transfer", transferOut),
			want:   map[int64]int64{ownerAcc: 100, intruderAcc: 100},
		},
		{
			name:   "transfer crediting another user's account",
			status: http.StatusForbidden,
			handle: app.transferHandler,
			req:    requestWithUser(intruder, "", "transfer", transferIn),
			want:   map[int64]int64{ownerAcc: 100, intruderAcc: 100},
		},
		{
			name:   "deposit into one's own account",
			status: http.StatusOK,
			handle: app.depositHandler,
			req:    requestWithUser(owner, ownerID, "deposit", `{"amount":50}`),
			want:   map[int64]int64{ownerAcc: 150},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.handle(w, tt.req)
			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d", w.Code, tt.status)
			}
			for id, want := range tt.want {
				var got int64
				if err := db.QueryRow("SELECT balance FROM accounts WHERE id = $1", id).Scan(&got); err != nil {
					t.Fatalf("read balance of account %d: %v", id, err)
				}
				if got != want {
					t.Fatalf("account %d balance = %d, want %d", id, got, want)
				}
			}
		})
	}
}

func requestWithUser(userID int64, id, action, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/v1/"+action, strings.NewReader(body))
	r.SetPathValue("id", id)
	return r.WithContext(
		context.WithValue(context.Background(), userContextKey, &user.User{ID: userID}),
	)
}
