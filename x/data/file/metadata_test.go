package file

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
	"time"
)

func TestNewDBMetadataManagerERejectsNilDB(t *testing.T) {
	m, err := NewDBMetadataManagerE(nil)
	if !errors.Is(err, ErrNilMetadataDB) {
		t.Fatalf("NewDBMetadataManagerE error = %v, want ErrNilMetadataDB", err)
	}
	if m != nil {
		t.Fatalf("manager = %#v, want nil", m)
	}
}

func TestDBMetadataManagerNilDBMethodsReturnSentinel(t *testing.T) {
	m := NewDBMetadataManager(nil).(*DBMetadataManager)
	ctx := t.Context()
	file := &File{ID: "f-1", Metadata: map[string]any{}}

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "Save", run: func() error { return m.Save(ctx, file) }},
		{name: "Get", run: func() error { _, err := m.Get(ctx, "f-1"); return err }},
		{name: "GetByPath", run: func() error { _, err := m.GetByPath(ctx, "tenant/f-1"); return err }},
		{name: "GetByHash", run: func() error { _, err := m.GetByHash(ctx, "hash"); return err }},
		{name: "List", run: func() error { _, _, err := m.List(ctx, Query{}); return err }},
		{name: "Delete", run: func() error { return m.Delete(ctx, "f-1") }},
		{name: "UpdateAccessTime", run: func() error { return m.UpdateAccessTime(ctx, "f-1") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); !errors.Is(err, ErrNilMetadataDB) {
				t.Fatalf("error = %v, want ErrNilMetadataDB", err)
			}
		})
	}
}

func TestDBMetadataManagerUsesConfiguredClockForMutations(t *testing.T) {
	rec := &metadataExecRecorder{}
	db := sql.OpenDB(metadataConnector{rec: rec})
	defer db.Close()

	fixed := time.Date(2026, 4, 25, 12, 30, 0, 0, time.UTC)
	m, err := NewDBMetadataManagerE(db, WithMetadataClock(func() time.Time { return fixed }))
	if err != nil {
		t.Fatalf("NewDBMetadataManagerE error = %v", err)
	}

	if err := m.Delete(t.Context(), "f-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := m.UpdateAccessTime(t.Context(), "f-1"); err != nil {
		t.Fatalf("UpdateAccessTime() error = %v", err)
	}

	if len(rec.args) != 2 {
		t.Fatalf("recorded exec count = %d, want 2", len(rec.args))
	}
	for i, args := range rec.args {
		if len(args) == 0 {
			t.Fatalf("exec %d recorded no args", i)
		}
		got, ok := args[0].Value.(time.Time)
		if !ok {
			t.Fatalf("exec %d first arg type = %T, want time.Time", i, args[0].Value)
		}
		if !got.Equal(fixed) {
			t.Fatalf("exec %d timestamp = %v, want %v", i, got, fixed)
		}
	}
}

func TestScanMetadataFileUnmarshalsMetadata(t *testing.T) {
	now := time.Date(2026, 4, 25, 13, 0, 0, 0, time.UTC)
	file, err := scanMetadataFile(func(dest ...any) error {
		values := []any{
			"f-1", "tenant-1", "avatar.png", "tenant-1/avatar.png", int64(12),
			"image/png", ".png", "hash", 10, 20,
			"tenant-1/thumb.png", "local", []byte(`{"kind":"avatar"}`),
			"user-1", now, now, (*time.Time)(nil), (*time.Time)(nil),
		}
		for i := range values {
			switch d := dest[i].(type) {
			case *string:
				*d = values[i].(string)
			case *int64:
				*d = values[i].(int64)
			case *int:
				*d = values[i].(int)
			case *[]byte:
				*d = values[i].([]byte)
			case *time.Time:
				*d = values[i].(time.Time)
			case **time.Time:
				*d = values[i].(*time.Time)
			default:
				t.Fatalf("unexpected destination type %T", d)
			}
		}
		return nil
	}, "test row")
	if err != nil {
		t.Fatalf("scanMetadataFile() error = %v", err)
	}
	if file.Metadata["kind"] != "avatar" {
		t.Fatalf("metadata kind = %v, want avatar", file.Metadata["kind"])
	}
}

type metadataExecRecorder struct {
	args [][]driver.NamedValue
}

type metadataConnector struct {
	rec *metadataExecRecorder
}

func (c metadataConnector) Connect(context.Context) (driver.Conn, error) {
	return metadataConn{rec: c.rec}, nil
}

func (c metadataConnector) Driver() driver.Driver {
	return metadataDriver{}
}

type metadataDriver struct{}

func (metadataDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("not used")
}

type metadataConn struct {
	rec *metadataExecRecorder
}

func (c metadataConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func (c metadataConn) Close() error {
	return nil
}

func (c metadataConn) Begin() (driver.Tx, error) {
	return nil, errors.New("not implemented")
}

func (c metadataConn) ExecContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Result, error) {
	copied := append([]driver.NamedValue(nil), args...)
	c.rec.args = append(c.rec.args, copied)
	return metadataResult(1), nil
}

func (c metadataConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return metadataRows{}, nil
}

type metadataResult int64

func (r metadataResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r metadataResult) RowsAffected() (int64, error) {
	return int64(r), nil
}

type metadataRows struct{}

func (metadataRows) Columns() []string {
	return nil
}

func (metadataRows) Close() error {
	return nil
}

func (metadataRows) Next([]driver.Value) error {
	return io.EOF
}
