package repository

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wyw/cry031-volunteer/internal/domain"
)

func TestPostgresStoreTransaction(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("cry031_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec(context.Background(), `DROP SCHEMA "`+schema+`" CASCADE`)

	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	testURL := parsed.String()
	pool, err := pgxpool.New(ctx, testURL)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"001_init.sql", "003_indexes.sql"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		for _, statement := range strings.Split(string(data), ";") {
			if strings.TrimSpace(statement) == "" {
				continue
			}
			if _, err := pool.Exec(ctx, statement); err != nil {
				t.Fatalf("migration %s: %v", name, err)
			}
		}
	}
	pool.Close()

	initial := domain.NewState()
	initial.Users["u1"] = domain.User{ID: "u1", DisplayName: "测试用户", Phone: "13900000000", Role: domain.RoleMember, Active: true, CreatedAt: time.Now().UTC()}
	store, err := NewPostgresStore(ctx, testURL, initial)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	state, err := store.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	state.Users["u2"] = domain.User{ID: "u2", DisplayName: "第二用户", Phone: "13900000001", Role: domain.RoleMember, Active: true, CreatedAt: time.Now().UTC()}
	event := domain.AuditEvent{ID: "audit-pg", ActorID: "u1", Action: "add_user", CreatedAt: time.Now().UTC()}
	state.Audit = append(state.Audit, event)
	if err := store.Commit(ctx, state.Version, state, []domain.AuditEvent{event}); err != nil {
		t.Fatal(err)
	}
	persisted, err := store.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := persisted.Users["u2"]; !ok || persisted.Version != 1 {
		t.Fatalf("persisted version=%d users=%v", persisted.Version, persisted.Users)
	}
}
