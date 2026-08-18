package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wyw/cry031-volunteer/internal/domain"
)

// PostgresStore persists the complete domain aggregate as a versioned JSONB snapshot and writes
// immutable audit rows in the same transaction. Relational projection tables are defined by migrations.
type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, databaseURL string, initial domain.State) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database configuration: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	store := &PostgresStore{pool: pool}
	data, err := json.Marshal(initial)
	if err != nil {
		pool.Close()
		return nil, err
	}
	_, err = pool.Exec(ctx, `INSERT INTO volunteer_state (singleton, version, payload) VALUES (TRUE, 0, $1) ON CONFLICT (singleton) DO NOTHING`, data)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("initialize volunteer state: %w", err)
	}
	return store, nil
}

func (s *PostgresStore) Close() { s.pool.Close() }

func (s *PostgresStore) Snapshot(ctx context.Context) (domain.State, error) {
	var version uint64
	var payload []byte
	if err := s.pool.QueryRow(ctx, `SELECT version, payload FROM volunteer_state WHERE singleton = TRUE`).Scan(&version, &payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.State{}, domain.ErrNotFound
		}
		return domain.State{}, fmt.Errorf("read volunteer state: %w", err)
	}
	var state domain.State
	if err := json.Unmarshal(payload, &state); err != nil {
		return domain.State{}, fmt.Errorf("decode volunteer state: %w", err)
	}
	state.Version = version
	return state, nil
}

func (s *PostgresStore) Commit(ctx context.Context, expected uint64, next domain.State, events []domain.AuditEvent) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var current uint64
	if err := tx.QueryRow(ctx, `SELECT version FROM volunteer_state WHERE singleton = TRUE FOR UPDATE`).Scan(&current); err != nil {
		return err
	}
	if current != expected {
		return domain.ErrConflict
	}
	next.Version = current + 1
	payload, err := json.Marshal(next)
	if err != nil {
		return err
	}
	result, err := tx.Exec(ctx, `UPDATE volunteer_state SET version = $1, payload = $2, updated_at = now() WHERE singleton = TRUE AND version = $3`, next.Version, payload, expected)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	for _, event := range events {
		if _, err := tx.Exec(ctx, `INSERT INTO audit_events (id, actor_id, entity_type, entity_id, action, details, request_id, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, event.ID, event.ActorID, event.EntityType, event.EntityID, event.Action, event.Details, event.RequestID, event.CreatedAt); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
