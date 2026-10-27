package account

import (
	"context"
	"testing"
	"time"

	"github.com/nongpal/Palge-Backend/internal/testutil"
)

func TestAccountInsert(t *testing.T) {
	db := testutil.NewTestDB(t)
	model := Model{DB: db}

	tests := []struct {
		name    string
		owner   string
		balance int64
		wantErr bool
	}{
		{
			name:    "valid account with positive balance",
			owner:   "test-owner-1",
			balance: 100,
			wantErr: false,
		},
		{
			name:    "valid account with zero balance",
			owner:   "test-owner-2",
			balance: 0,
			wantErr: false,
		},
		{
			name:    "invalid account with negative balance",
			owner:   "test-owner-3",
			balance: -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			acc := Account{
				Owner:   tt.owner,
				Balance: tt.balance,
			}

			err := model.Insert(ctx, &acc)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"Inser() err = %v, wantErr %v",
					err, tt.wantErr,
				)
			}

			if tt.wantErr {
				return
			}

			if acc.ID <= 0 {
				t.Fatalf("Insert() did not set account ID: %d", acc.ID)
			}

			t.Cleanup(func() {
				cleanupCtx, cleanupCancel := context.WithTimeout(
					context.Background(),
					5*time.Second,
				)
				defer cleanupCancel()

				_, err := db.ExecContext(
					cleanupCtx,
					"DELETE FROM accounts WHERE id = $1", acc.ID,
				)

				if err != nil {
					t.Fatalf("cleanup account %d: %v", acc.ID, err)
				}
			})

			var got Account

			err = db.QueryRowContext(
				ctx,
				`
				SELECT id, owner, balance
				FROM accounts
				WHERE id = $1
				`,
				acc.ID,
			).Scan(
				&got.ID,
				&got.Owner,
				&got.Balance,
			)

			if err != nil {
				t.Fatalf("query inserted account: %v", err)
			}

			if got.ID != acc.ID {
				t.Errorf("ID = %d, want %d", got.ID, acc.ID)
			}

			if got.Owner != tt.owner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.owner)
			}

			if got.Balance != tt.balance {
				t.Errorf("Balance = %d, want %d", got.Balance, tt.balance)
			}
		})
	}
}

func TestGetAccount(t *testing.T) {
	db := testutil.NewTestDB(t)
	model := Model{DB: db}

	tests := []struct {
		name            string
		id              int64
		seedAccount     bool
		deleteBeforeGet bool
		owner           string
		balance         int64
		wantErr         bool
	}{
		{
			name:        "get existing account",
			seedAccount: true,
			owner:       "test-owner-1",
			balance:     100,
			wantErr:     false,
		},
		{
			name:    "get account with ID less than 1",
			id:      0,
			wantErr: true,
		},
		{
			name:            "get account that does not exist",
			seedAccount:     true,
			deleteBeforeGet: true,
			owner:           "test-owner-2",
			balance:         200,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			id := tt.id

			var inserted Account

			if tt.seedAccount {
				inserted = Account{
					Owner:   tt.owner,
					Balance: tt.balance,
				}

				err := model.Insert(ctx, &inserted)
				if err != nil {
					t.Fatalf(
						"Insert test account: %v",
						err,
					)
				}

				id = inserted.ID
				accountID := inserted.ID

				t.Cleanup(func() {
					cleanupCtx, cleanupCancel := context.WithTimeout(
						context.Background(),
						5*time.Second,
					)
					defer cleanupCancel()

					_, err := db.ExecContext(
						cleanupCtx,
						"DELETE FROM accounts WHERE id = $1",
						accountID,
					)
					if err != nil {
						t.Errorf("cleanup account: %v", err)
					}
				})

				if tt.deleteBeforeGet {
					_, err := db.ExecContext(ctx,
						`DELETE FROM accounts WHERE id = $1`,
						id,
					)

					if err != nil {
						t.Fatalf("delete test account: %v", err)
					}
				}
			}

			got, err := model.Get(ctx, id)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"Got(%d) error = %v, wantErr %v",
					id, err, tt.wantErr,
				)
			}

			if tt.wantErr {
				if got != nil {
					t.Errorf("Got() = %+v, want nil account", *got)
				}
				return
			}

			if got == nil {
				t.Fatal("Get() returned nil account without an error")
			}

			if got.ID != inserted.ID {
				t.Errorf("ID = %d, want %d", got.ID, inserted.ID)
			}

			if got.Owner != tt.owner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.owner)
			}

			if got.Balance != tt.balance {
				t.Errorf("Balance = %d, want %d", got.Balance, tt.balance)
			}
		})
	}
}
