package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "missing driver",
			config:  Config{DSN: "dsn"},
			wantErr: true,
		},
		{
			name:    "missing dsn",
			config:  Config{Driver: "driver"},
			wantErr: true,
		},
		{
			name:    "negative max open",
			config:  Config{Driver: "driver", DSN: "dsn", MaxOpenConns: -1},
			wantErr: true,
		},
		{
			name:    "negative ping timeout",
			config:  Config{Driver: "driver", DSN: "dsn", PingTimeout: -1},
			wantErr: true,
		},
		{
			name:    "valid",
			config:  Config{Driver: "driver", DSN: "dsn", MaxOpenConns: 5},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		if err := tc.config.Validate(); (err != nil) != tc.wantErr {
			t.Fatalf("%s: Validate() error = %v", tc.name, err)
		}
	}
}

func TestApplyConfigMaxOpenConns(t *testing.T) {
	conn := &stubConn{}
	connector := &stubConnector{conn: conn}
	db := sql.OpenDB(connector)
	defer db.Close()

	ApplyConfig(db, Config{MaxOpenConns: 4, MaxIdleConns: 10})
	stats := db.Stats()
	if stats.MaxOpenConnections != 4 {
		t.Fatalf("expected MaxOpenConnections 4, got %d", stats.MaxOpenConnections)
	}
}

func TestOpenWithPing(t *testing.T) {
	pingErr := errors.New("ping failed")
	conn := &stubConn{pingErr: pingErr}
	connector := &stubConnector{conn: conn}

	_, err := OpenWith(Config{
		Driver:      "stub",
		DSN:         "dsn",
		PingTimeout: 50 * time.Millisecond,
	}, func(driver, dsn string) (*sql.DB, error) {
		return sql.OpenDB(connector), nil
	})
	if err == nil || !errors.Is(err, pingErr) {
		t.Fatalf("expected ping error, got %v", err)
	}
}

type stubConnector struct {
	conn *stubConn
}

func (c *stubConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.conn, nil
}

func (c *stubConnector) Driver() driver.Driver {
	return stubDriver{}
}

type stubDriver struct{}

func (d stubDriver) Open(name string) (driver.Conn, error) {
	return nil, errors.New("not supported")
}

type stubConn struct {
	pingErr error
}

func (c *stubConn) Prepare(query string) (driver.Stmt, error) {
	return stubStmt{}, nil
}

func (c *stubConn) Close() error {
	return nil
}

func (c *stubConn) Begin() (driver.Tx, error) {
	return stubTx{}, nil
}

func (c *stubConn) Ping(ctx context.Context) error {
	return c.pingErr
}

type stubStmt struct{}

func (s stubStmt) Close() error {
	return nil
}

func (s stubStmt) NumInput() int {
	return -1
}

func (s stubStmt) Exec(args []driver.Value) (driver.Result, error) {
	return stubResult{}, nil
}

func (s stubStmt) Query(args []driver.Value) (driver.Rows, error) {
	return stubRows{}, nil
}

type stubTx struct{}

func (t stubTx) Commit() error {
	return nil
}

func (t stubTx) Rollback() error {
	return nil
}

type stubResult struct{}

func (r stubResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r stubResult) RowsAffected() (int64, error) {
	return 0, nil
}

type stubRows struct{}

func (r stubRows) Columns() []string {
	return nil
}

func (r stubRows) Close() error {
	return nil
}

func (r stubRows) Next(dest []driver.Value) error {
	return io.EOF
}
