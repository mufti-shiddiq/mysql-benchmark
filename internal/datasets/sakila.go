package datasets

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mufti-shiddiq/mysql-benchmark/internal/benchmark"
)

const sakilaPrefix = "benchmark_sakila_"

func PrepareSakila(ctx context.Context, db *sql.DB, force bool) error {
	if force {
		if err := dropPrefixedTables(ctx, db, sakilaPrefix); err != nil {
			return err
		}
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS benchmark_sakila_category (category_id INT PRIMARY KEY, name VARCHAR(64) NOT NULL) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_sakila_film (film_id INT PRIMARY KEY, title VARCHAR(128) NOT NULL, rental_rate DECIMAL(5,2) NOT NULL, length_minutes INT NOT NULL, INDEX idx_title (title), INDEX idx_rate (rental_rate)) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_sakila_film_category (film_id INT NOT NULL, category_id INT NOT NULL, PRIMARY KEY (film_id, category_id), INDEX idx_category (category_id)) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_sakila_customer (customer_id INT PRIMARY KEY, store_id INT NOT NULL, first_name VARCHAR(64) NOT NULL, last_name VARCHAR(64) NOT NULL, email VARCHAR(128) NOT NULL, active BOOLEAN NOT NULL, INDEX idx_store_active (store_id, active), INDEX idx_last_name (last_name)) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_sakila_inventory (inventory_id INT PRIMARY KEY, film_id INT NOT NULL, store_id INT NOT NULL, INDEX idx_film_store (film_id, store_id)) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_sakila_rental (rental_id INT PRIMARY KEY, rental_date DATETIME NOT NULL, inventory_id INT NOT NULL, customer_id INT NOT NULL, return_date DATETIME NULL, INDEX idx_customer (customer_id), INDEX idx_inventory (inventory_id), INDEX idx_rental_date (rental_date)) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_sakila_payment (payment_id INT PRIMARY KEY, customer_id INT NOT NULL, rental_id INT NOT NULL, amount DECIMAL(6,2) NOT NULL, payment_date DATETIME NOT NULL, INDEX idx_customer_date (customer_id, payment_date), INDEX idx_rental (rental_id)) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_sakila_write (id BIGINT PRIMARY KEY AUTO_INCREMENT, note VARCHAR(128) NOT NULL, amount DECIMAL(8,2) NOT NULL, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP) ENGINE=InnoDB`,
	}
	if err := execMany(ctx, db, statements); err != nil {
		return err
	}
	count, err := tableCount(ctx, db, "benchmark_sakila_customer")
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return seedSakila(ctx, db)
}

func seedSakila(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i := 1; i <= 16; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_sakila_category(category_id, name) VALUES (?, ?)`, i, fmt.Sprintf("Category %02d", i)); err != nil {
			return err
		}
	}
	for i := 1; i <= 1000; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_sakila_film(film_id, title, rental_rate, length_minutes) VALUES (?, ?, ?, ?)`, i, fmt.Sprintf("Film %04d", i), float64((i%7)+1)+0.99, 80+(i%90)); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_sakila_film_category(film_id, category_id) VALUES (?, ?)`, i, (i%16)+1); err != nil {
			return err
		}
	}
	for i := 1; i <= 800; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_sakila_customer(customer_id, store_id, first_name, last_name, email, active) VALUES (?, ?, ?, ?, ?, ?)`, i, (i%2)+1, fmt.Sprintf("First%03d", i), fmt.Sprintf("Last%03d", i%200), fmt.Sprintf("customer%03d@example.test", i), i%13 != 0); err != nil {
			return err
		}
	}
	for i := 1; i <= 3000; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_sakila_inventory(inventory_id, film_id, store_id) VALUES (?, ?, ?)`, i, (i%1000)+1, (i%2)+1); err != nil {
			return err
		}
	}
	for i := 1; i <= 6000; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_sakila_rental(rental_id, rental_date, inventory_id, customer_id, return_date) VALUES (?, DATE_ADD('2026-01-01', INTERVAL ? HOUR), ?, ?, DATE_ADD('2026-01-03', INTERVAL ? HOUR))`, i, i%2000, (i%3000)+1, (i%800)+1, i%2000); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_sakila_payment(payment_id, customer_id, rental_id, amount, payment_date) VALUES (?, ?, ?, ?, DATE_ADD('2026-01-02', INTERVAL ? HOUR))`, i, (i%800)+1, i, float64((i%9)+1)+0.49, i%2000); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func SakilaCases(db *sql.DB) []benchmark.Case {
	return []benchmark.Case{
		{Name: "select_1_latency", Description: "Round trip plus MySQL processing overhead", Run: func(ctx context.Context) error { return runQuery(ctx, db, "SELECT 1") }},
		{Name: "simple_select", Description: "Indexed customer lookup", Run: func(ctx context.Context) error {
			return runQuery(ctx, db, "SELECT customer_id, email FROM benchmark_sakila_customer WHERE store_id = ? AND active = ? LIMIT 50", 1, true)
		}},
		{Name: "two_table_join", Description: "Customer to rental join", Run: func(ctx context.Context) error {
			return runQuery(ctx, db, `SELECT c.customer_id, COUNT(r.rental_id) FROM benchmark_sakila_customer c JOIN benchmark_sakila_rental r ON r.customer_id = c.customer_id WHERE c.active = 1 GROUP BY c.customer_id LIMIT 100`)
		}},
		{Name: "multi_table_join", Description: "Customer, rental, inventory, film", Run: func(ctx context.Context) error {
			return runQuery(ctx, db, `SELECT c.customer_id, f.title FROM benchmark_sakila_customer c JOIN benchmark_sakila_rental r ON r.customer_id = c.customer_id JOIN benchmark_sakila_inventory i ON i.inventory_id = r.inventory_id JOIN benchmark_sakila_film f ON f.film_id = i.film_id WHERE c.store_id = 1 ORDER BY r.rental_date DESC LIMIT 100`)
		}},
		{Name: "complex_join", Description: "Customer rental category path", Run: func(ctx context.Context) error {
			return runQuery(ctx, db, `SELECT c.customer_id, cat.name, SUM(p.amount) FROM benchmark_sakila_customer c JOIN benchmark_sakila_rental r ON r.customer_id = c.customer_id JOIN benchmark_sakila_inventory i ON i.inventory_id = r.inventory_id JOIN benchmark_sakila_film f ON f.film_id = i.film_id JOIN benchmark_sakila_film_category fc ON fc.film_id = f.film_id JOIN benchmark_sakila_category cat ON cat.category_id = fc.category_id JOIN benchmark_sakila_payment p ON p.rental_id = r.rental_id GROUP BY c.customer_id, cat.name ORDER BY SUM(p.amount) DESC LIMIT 100`)
		}},
		{Name: "aggregation", Description: "COUNT SUM AVG GROUP BY", Run: func(ctx context.Context) error {
			return runQuery(ctx, db, `SELECT customer_id, COUNT(*), SUM(amount), AVG(amount) FROM benchmark_sakila_payment GROUP BY customer_id ORDER BY SUM(amount) DESC LIMIT 100`)
		}},
		{Name: "sorting", Description: "ORDER BY with LIMIT", Run: func(ctx context.Context) error {
			return runQuery(ctx, db, `SELECT film_id, title, rental_rate FROM benchmark_sakila_film ORDER BY rental_rate DESC, length_minutes DESC LIMIT 100`)
		}},
		{Name: "subquery", Description: "Customers above average payment", Run: func(ctx context.Context) error {
			return runQuery(ctx, db, `SELECT customer_id, SUM(amount) total FROM benchmark_sakila_payment GROUP BY customer_id HAVING total > (SELECT AVG(amount) * 8 FROM benchmark_sakila_payment) LIMIT 100`)
		}},
		{Name: "insert", Description: "Controlled insert", Run: func(ctx context.Context) error {
			_, err := db.ExecContext(ctx, `INSERT INTO benchmark_sakila_write(note, amount) VALUES ('insert benchmark', 1.23)`)
			return err
		}},
		{Name: "update", Description: "Controlled update", Run: func(ctx context.Context) error {
			_, err := db.ExecContext(ctx, `UPDATE benchmark_sakila_write SET amount = amount + 1 WHERE id = (SELECT id FROM (SELECT MAX(id) id FROM benchmark_sakila_write) x)`)
			return err
		}},
		{Name: "delete", Description: "Controlled delete", Run: func(ctx context.Context) error {
			_, err := db.ExecContext(ctx, `DELETE FROM benchmark_sakila_write WHERE id = (SELECT id FROM (SELECT MIN(id) id FROM benchmark_sakila_write) x)`)
			return err
		}},
		{Name: "transaction", Description: "BEGIN INSERT UPDATE SELECT COMMIT", Run: sakilaTransaction(db)},
	}
}

func sakilaTransaction(db *sql.DB) func(context.Context) error {
	return func(ctx context.Context) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		res, err := tx.ExecContext(ctx, `INSERT INTO benchmark_sakila_write(note, amount) VALUES ('transaction benchmark', 2.34)`)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		if _, err := tx.ExecContext(ctx, `UPDATE benchmark_sakila_write SET amount = amount + 1 WHERE id = ?`, id); err != nil {
			return err
		}
		var amount float64
		if err := tx.QueryRowContext(ctx, `SELECT amount FROM benchmark_sakila_write WHERE id = ?`, id).Scan(&amount); err != nil {
			return err
		}
		return tx.Commit()
	}
}
