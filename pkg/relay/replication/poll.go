package replication

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/controlplane-com/libs-go/pkg/logging"
)

type ChangeHandler func(*Change) error

func startPolling(c *Config) error {
	conn, err := connection(c)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())
	logging.Logger().Sugar().Infof("Outbox polling started for tables %v every %s", c.Tables(), c.PollInterval)
	for {
		if err := Poll(context.Background(), conn, c.Tables(), c.PollBatchSize, c.Destinations.HandleChange); err != nil {
			return err
		}
		time.Sleep(c.PollInterval)
	}
}

// Poll delivers each table's undelivered outbox rows to handle until a short batch, marking each delivered.
func Poll(ctx context.Context, conn *pgx.Conn, tables []string, batchSize int, handle ChangeHandler) error {
	if batchSize <= 0 {
		return fmt.Errorf("invalid poll batch size: %d", batchSize)
	}
	for _, table := range tables {
		for {
			n, err := pollBatch(ctx, conn, table, batchSize, handle)
			if err != nil {
				return err
			}
			if n < batchSize {
				break
			}
		}
	}
	return nil
}

// pollBatch returns the number of rows delivered; a handler error ends the batch early and is retried next poll.
func pollBatch(ctx context.Context, conn *pgx.Conn, table string, batchSize int, handle ChangeHandler) (int, error) {
	schema, name, ok := strings.Cut(table, ".")
	if !ok {
		return 0, fmt.Errorf("outbox table %q must be schema-qualified", table)
	}
	ident := pgx.Identifier{schema, name}.Sanitize()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, fmt.Sprintf(
		"SELECT id::text, created::text, payload::text FROM %s WHERE delivered IS NULL ORDER BY created, id LIMIT $1 FOR UPDATE SKIP LOCKED", ident), batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to read outbox %s: %w", table, err)
	}
	var changes []*Change
	for rows.Next() {
		var id, created, payload string
		if err := rows.Scan(&id, &created, &payload); err != nil {
			rows.Close()
			return 0, fmt.Errorf("failed to scan outbox %s: %w", table, err)
		}
		changes = append(changes, &Change{
			Action:       ActionInsert,
			Schema:       schema,
			Table:        name,
			ColumnNames:  []string{"id", "created", "payload"},
			ColumnValues: []any{id, created, payload},
		})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("failed to read outbox %s: %w", table, err)
	}

	delivered := 0
	for _, change := range changes {
		if err := handle(change); err != nil && !errors.Is(err, IgnoreMessageError) {
			logging.Logger().Sugar().Errorf("error delivering outbox %s row %v, retrying next poll: %v", table, change.ColumnValues[0], err)
			break
		}
		if _, err := tx.Exec(ctx, fmt.Sprintf("UPDATE %s SET delivered = now() WHERE id = $1", ident), change.ColumnValues[0]); err != nil {
			return 0, fmt.Errorf("failed to mark outbox %s row delivered: %w", table, err)
		}
		delivered++
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit outbox %s: %w", table, err)
	}
	if delivered < len(changes) {
		return 0, nil
	}
	return delivered, nil
}
