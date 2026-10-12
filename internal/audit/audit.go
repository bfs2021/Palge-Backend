package audit

import (
	"context"
	"database/sql"
	"time"
)

type AuditLog struct {
	ID         int64     `json:"id"`
	UserID     *int64    `json:"user_id"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID *int64    `json:"resource_id"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Result     string    `json:"result"`
	CreatedAt  time.Time `json:"created_at"`
}

type Model struct {
	DB *sql.DB
}

func (m *Model) Insert(auditLog *AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			user_id,
			action,
			resource,
			resource_id,
			ip_address,
			user_agent,
			result
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(
		ctx,
		query,
		auditLog.UserID,
		auditLog.Action,
		auditLog.Resource,
		auditLog.ResourceID,
		auditLog.IPAddress,
		auditLog.UserAgent,
		auditLog.Result,
	).Scan(
		&auditLog.ID,
		&auditLog.CreatedAt,
	)
}
