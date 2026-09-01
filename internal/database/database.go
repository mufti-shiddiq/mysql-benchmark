package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strings"
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
	mysqlCfg.Params = map[string]string{"charset": "utf8mb4"}

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
		return info, friendlyConnectionError(err)
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

func friendlyConnectionError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1045:
			return errors.New("access denied. Please verify username, password, and whether the MySQL user is allowed to connect from this host")
		case 1049:
			return errors.New("database does not exist. Create the database first or choose an existing database")
		case 2002, 2003:
			return errors.New("unable to reach MySQL. Please verify host, port, bind address, firewall rules, and that MySQL is running")
		}
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return errors.New("unable to connect to MySQL: connection timed out")
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "connection refused"):
		return errors.New("unable to connect to MySQL: connection refused. Please verify MySQL is running and listening on the selected host and port")
	case strings.Contains(msg, "operation not permitted"):
		return errors.New("unable to connect to MySQL: operation not permitted. If running inside a sandboxed terminal, allow local network access or run the binary from a normal shell")
	default:
		return fmt.Errorf("unable to connect to MySQL: %w", err)
	}
}

func TableCount(ctx context.Context, db *sql.DB, database string) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = ?", database).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unable to inspect database tables: %w", err)
	}
	return count, nil
}
