package message

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spcent/plumego/net/mq"
)

type SQLDialect string

const (
	DialectPostgres SQLDialect = "postgres"
	DialectMySQL    SQLDialect = "mysql"
)

type SQLConfig struct {
	Dialect SQLDialect
	Table   string
	Now     func() time.Time
}

func DefaultSQLConfig() SQLConfig {
	return SQLConfig{
		Dialect: DialectPostgres,
		Table:   "sms_messages",
		Now:     time.Now,
	}
}

// SQLExecutor abstracts the minimal database operations used by SQLRepository.
type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// SQLRepository is a SQL-backed Store implementation.
type SQLRepository struct {
	db  SQLExecutor
	cfg SQLConfig
	now func() time.Time
}

func NewSQLRepository(db SQLExecutor, cfg SQLConfig) *SQLRepository {
	if cfg.Table == "" {
		cfg.Table = "sms_messages"
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Dialect == "" {
		cfg.Dialect = DialectPostgres
	}
	return &SQLRepository{db: db, cfg: cfg, now: cfg.Now}
}

func (r *SQLRepository) Insert(ctx context.Context, msg Message) error {
	if r == nil || r.db == nil {
		return ErrMessageNotFound
	}
	if strings.TrimSpace(msg.ID) == "" {
		return ErrMessageNotFound
	}

	now := r.now()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	msg.UpdatedAt = now
	if msg.Status == "" {
		msg.Status = StatusAccepted
	}
	if msg.MaxAttempts <= 0 {
		msg.MaxAttempts = mq.DefaultTaskMaxAttempts
	}

	query, args := r.buildInsert(msg)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		if isDuplicateError(err) {
			return ErrMessageExists
		}
		return err
	}
	return nil
}

func (r *SQLRepository) Get(ctx context.Context, id string) (Message, bool, error) {
	if r == nil || r.db == nil {
		return Message{}, false, ErrMessageNotFound
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return Message{}, false, ErrMessageNotFound
	}

	query := fmt.Sprintf(`SELECT id, tenant_id, provider, status, reason_code, reason_detail,
    attempts, max_attempts, next_attempt_at, sent_at, provider_msg_id, idempotency_key, version, created_at, updated_at
    FROM %s WHERE id = %s`, r.cfg.Table, r.placeholder(1))
	row := r.db.QueryRowContext(ctx, query, id)

	var msg Message
	var provider sql.NullString
	var status string
	var reasonCode sql.NullString
	var reasonDetail sql.NullString
	var nextAttempt sql.NullTime
	var sentAt sql.NullTime
	var providerMsgID sql.NullString
	var idempotencyKey sql.NullString
	if err := row.Scan(
		&msg.ID,
		&msg.TenantID,
		&provider,
		&status,
		&reasonCode,
		&reasonDetail,
		&msg.Attempts,
		&msg.MaxAttempts,
		&nextAttempt,
		&sentAt,
		&providerMsgID,
		&idempotencyKey,
		&msg.Version,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Message{}, false, nil
		}
		return Message{}, false, err
	}

	msg.Status = Status(status)
	if provider.Valid {
		msg.Provider = provider.String
	}
	if reasonCode.Valid {
		msg.ReasonCode = ReasonCode(reasonCode.String)
	}
	if reasonDetail.Valid {
		msg.ReasonDetail = reasonDetail.String
	}
	if nextAttempt.Valid {
		msg.NextAttemptAt = nextAttempt.Time
	}
	if sentAt.Valid {
		msg.SentAt = sentAt.Time
	}
	if providerMsgID.Valid {
		msg.ProviderMsgID = providerMsgID.String
	}
	if idempotencyKey.Valid {
		msg.IdempotencyKey = idempotencyKey.String
	}

	return msg, true, nil
}

func (r *SQLRepository) UpdateStatus(ctx context.Context, id string, from Status, to Status, reason Reason) error {
	if r == nil || r.db == nil {
		return ErrMessageNotFound
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrMessageNotFound
	}

	now := r.now()
	var res sql.Result
	var err error
	if to == StatusSent {
		query := fmt.Sprintf(`UPDATE %s SET status = %s, reason_code = %s, reason_detail = %s,
    sent_at = %s, updated_at = %s, version = version + 1 WHERE id = %s AND status = %s`,
			r.cfg.Table,
			r.placeholder(1),
			r.placeholder(2),
			r.placeholder(3),
			r.placeholder(4),
			r.placeholder(5),
			r.placeholder(6),
			r.placeholder(7),
		)
		res, err = r.db.ExecContext(ctx, query, string(to), nullString(string(reason.Code)), nullString(reason.Detail), now, now, id, string(from))
	} else {
		query := fmt.Sprintf(`UPDATE %s SET status = %s, reason_code = %s, reason_detail = %s,
    updated_at = %s, version = version + 1 WHERE id = %s AND status = %s`,
			r.cfg.Table,
			r.placeholder(1),
			r.placeholder(2),
			r.placeholder(3),
			r.placeholder(4),
			r.placeholder(5),
			r.placeholder(6),
		)
		res, err = r.db.ExecContext(ctx, query, string(to), nullString(string(reason.Code)), nullString(reason.Detail), now, id, string(from))
	}
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return r.resolveUpdateConflict(ctx, id)
	}
	return nil
}

func (r *SQLRepository) UpdateProvider(ctx context.Context, id string, provider string) error {
	if r == nil || r.db == nil {
		return ErrMessageNotFound
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrMessageNotFound
	}

	now := r.now()
	query := fmt.Sprintf("UPDATE %s SET provider = %s, updated_at = %s, version = version + 1 WHERE id = %s",
		r.cfg.Table,
		r.placeholder(1),
		r.placeholder(2),
		r.placeholder(3),
	)
	res, err := r.db.ExecContext(ctx, query, nullString(provider), now, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrMessageNotFound
	}
	return nil
}

func (r *SQLRepository) RecordAttempt(ctx context.Context, id string, attempts int) error {
	if r == nil || r.db == nil {
		return ErrMessageNotFound
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrMessageNotFound
	}

	now := r.now()
	query := fmt.Sprintf("UPDATE %s SET attempts = %s, updated_at = %s, version = version + 1 WHERE id = %s",
		r.cfg.Table,
		r.placeholder(1),
		r.placeholder(2),
		r.placeholder(3),
	)
	res, err := r.db.ExecContext(ctx, query, attempts, now, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrMessageNotFound
	}
	return nil
}

func (r *SQLRepository) RecordDLQ(ctx context.Context, id string, reason Reason) error {
	if r == nil || r.db == nil {
		return ErrMessageNotFound
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrMessageNotFound
	}

	now := r.now()
	query := fmt.Sprintf(`UPDATE %s SET status = %s, reason_code = %s, reason_detail = %s,
    updated_at = %s, version = version + 1 WHERE id = %s`,
		r.cfg.Table,
		r.placeholder(1),
		r.placeholder(2),
		r.placeholder(3),
		r.placeholder(4),
		r.placeholder(5),
	)
	res, err := r.db.ExecContext(ctx, query, string(StatusFailed), nullString(string(reason.Code)), nullString(reason.Detail), now, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrMessageNotFound
	}
	return nil
}

func (r *SQLRepository) resolveUpdateConflict(ctx context.Context, id string) error {
	query := fmt.Sprintf("SELECT status FROM %s WHERE id = %s", r.cfg.Table, r.placeholder(1))
	row := r.db.QueryRowContext(ctx, query, id)
	var status string
	if err := row.Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrMessageNotFound
		}
		return err
	}
	return ErrMessageStateConflict
}

func (r *SQLRepository) buildInsert(msg Message) (string, []any) {
	cols := []string{
		"id",
		"tenant_id",
		"provider",
		"status",
		"reason_code",
		"reason_detail",
		"attempts",
		"max_attempts",
		"next_attempt_at",
		"sent_at",
		"provider_msg_id",
		"idempotency_key",
		"version",
		"created_at",
		"updated_at",
	}

	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = r.placeholder(i + 1)
	}

	args := []any{
		msg.ID,
		msg.TenantID,
		nullString(msg.Provider),
		string(msg.Status),
		nullString(string(msg.ReasonCode)),
		nullString(msg.ReasonDetail),
		msg.Attempts,
		msg.MaxAttempts,
		nullTime(msg.NextAttemptAt),
		nullTime(msg.SentAt),
		nullString(msg.ProviderMsgID),
		nullString(msg.IdempotencyKey),
		msg.Version,
		msg.CreatedAt,
		msg.UpdatedAt,
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", r.cfg.Table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	return query, args
}

func (r *SQLRepository) placeholder(idx int) string {
	if r.cfg.Dialect == DialectPostgres {
		return fmt.Sprintf("$%d", idx)
	}
	return "?"
}

func nullString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique") || strings.Contains(msg, "constraint")
}
