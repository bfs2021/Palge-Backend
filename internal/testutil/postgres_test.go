package testutil

import (
	"context"
	"testing"
	"time"
)

func TestNewTestDB(t *testing.T) {
	db := NewTestDB(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	var got int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&got); err != nil {
		t.Fatalf("query SELECT 1: %v", err)
	}

	const want = 1
	if got != want {
		t.Errorf("SELECT 1 = %d, want %d", got, want)
	}
}
