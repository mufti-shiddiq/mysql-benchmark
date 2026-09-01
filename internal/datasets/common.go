package datasets

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func execMany(ctx context.Context, db *sql.DB, statements []string) error {
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("statement failed: %w", err)
		}
	}
	return nil
}

func tableCount(ctx context.Context, db *sql.DB, table string) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func dropPrefixedTables(ctx context.Context, db *sql.DB, prefix string) error {
	rows, err := db.QueryContext(ctx, "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME LIKE ?", prefix+"%")
	if err != nil {
		return err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return err
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(tables) == 0 {
		return nil
	}
	if _, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
		return err
	}
	defer db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=1")
	for _, table := range tables {
		quoted := "`" + strings.ReplaceAll(table, "`", "``") + "`"
		if _, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS "+quoted); err != nil {
			return err
		}
	}
	return nil
}

func runQuery(ctx context.Context, db *sql.DB, query string, args ...any) error {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
	}
	return rows.Err()
}
