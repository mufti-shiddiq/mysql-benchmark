package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/mufti-shiddiq/mysql-benchmark/internal/config"
)

type Info struct {
	Host         string
	Port         int
	Database     string
	MySQLVersion string
	TableCount   int
}

func Open(cfg config.Config) (*sql.DB, error) {
	mysqlCfg := mysql.NewConfig()
	mysqlCfg.User = cfg.User
	mysqlCfg.Passwd = cfg.Password
	mysqlCfg.Net = "tcp"
	mysqlCfg.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	mysqlCfg.DBName = cfg.Database
	mysqlCfg.ParseTime = true
	mysqlCfg.Timeout = 10 * time.Second
	mysqlCfg.ReadTimeout = 30 * time.Second
	mysqlCfg.WriteTimeout = 30 * time.Second
	mysqlCfg.Params = map[string]string{"charset": "utf8mb4,utf8"}

	db, err := sql.Open("mysql", mysqlCfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}

func Validate(ctx context.Context, db *sql.DB, cfg config.Config) (Info, error) {
	info := Info{Host: cfg.Host, Port: cfg.Port, Database: cfg.Database}
	if err := db.PingContext(ctx); err != nil {
		return info, fmt.Errorf("unable to connect to MySQL: %w", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&info.MySQLVersion); err != nil {
		return info, fmt.Errorf("unable to read MySQL version: %w", err)
	}
	var exists int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?", cfg.Database).Scan(&exists); err != nil {
		return info, fmt.Errorf("unable to verify database: %w", err)
	}
	if exists == 0 {
		return info, fmt.Errorf("database %q does not exist", cfg.Database)
	}
	count, err := TableCount(ctx, db, cfg.Database)
	if err != nil {
		return info, err
	}
	info.TableCount = count
	return info, nil
}

func TableCount(ctx context.Context, db *sql.DB, database string) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = ?", database).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unable to inspect database tables: %w", err)
	}
	return count, nil
}
