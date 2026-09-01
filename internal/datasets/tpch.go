package datasets

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mufti-shiddiq/mysql-benchmark/internal/benchmark"
)

const tpchPrefix = "benchmark_tpch_"

func PrepareTPCH(ctx context.Context, db *sql.DB, force bool, scaleFactor int) error {
	if force {
		if err := dropPrefixedTables(ctx, db, tpchPrefix); err != nil {
			return err
		}
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS benchmark_tpch_region (r_regionkey INT PRIMARY KEY, r_name VARCHAR(32) NOT NULL) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_tpch_nation (n_nationkey INT PRIMARY KEY, n_name VARCHAR(32) NOT NULL, n_regionkey INT NOT NULL, INDEX idx_region (n_regionkey)) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_tpch_supplier (s_suppkey INT PRIMARY KEY, s_name VARCHAR(64) NOT NULL, s_nationkey INT NOT NULL, s_acctbal DECIMAL(12,2) NOT NULL, INDEX idx_nation (s_nationkey)) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_tpch_customer (c_custkey INT PRIMARY KEY, c_name VARCHAR(64) NOT NULL, c_nationkey INT NOT NULL, c_acctbal DECIMAL(12,2) NOT NULL, c_mktsegment VARCHAR(32) NOT NULL, INDEX idx_nation (c_nationkey), INDEX idx_segment (c_mktsegment)) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_tpch_part (p_partkey INT PRIMARY KEY, p_name VARCHAR(128) NOT NULL, p_mfgr VARCHAR(32) NOT NULL, p_brand VARCHAR(32) NOT NULL, p_type VARCHAR(64) NOT NULL, p_size INT NOT NULL, p_retailprice DECIMAL(12,2) NOT NULL, INDEX idx_brand_type (p_brand, p_type), INDEX idx_size (p_size)) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_tpch_partsupp (ps_partkey INT NOT NULL, ps_suppkey INT NOT NULL, ps_availqty INT NOT NULL, ps_supplycost DECIMAL(12,2) NOT NULL, PRIMARY KEY (ps_partkey, ps_suppkey), INDEX idx_supp (ps_suppkey)) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_tpch_orders (o_orderkey INT PRIMARY KEY, o_custkey INT NOT NULL, o_orderstatus CHAR(1) NOT NULL, o_totalprice DECIMAL(12,2) NOT NULL, o_orderdate DATE NOT NULL, o_orderpriority VARCHAR(32) NOT NULL, INDEX idx_cust (o_custkey), INDEX idx_date (o_orderdate), INDEX idx_priority (o_orderpriority)) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS benchmark_tpch_lineitem (l_orderkey INT NOT NULL, l_partkey INT NOT NULL, l_suppkey INT NOT NULL, l_linenumber INT NOT NULL, l_quantity DECIMAL(12,2) NOT NULL, l_extendedprice DECIMAL(12,2) NOT NULL, l_discount DECIMAL(5,2) NOT NULL, l_tax DECIMAL(5,2) NOT NULL, l_returnflag CHAR(1) NOT NULL, l_linestatus CHAR(1) NOT NULL, l_shipdate DATE NOT NULL, l_commitdate DATE NOT NULL, l_receiptdate DATE NOT NULL, l_shipmode VARCHAR(16) NOT NULL, PRIMARY KEY (l_orderkey, l_linenumber), INDEX idx_shipdate (l_shipdate), INDEX idx_part_supp (l_partkey, l_suppkey), INDEX idx_supp (l_suppkey)) ENGINE=InnoDB`,
	}
	if err := execMany(ctx, db, statements); err != nil {
		return err
	}
	count, err := tableCount(ctx, db, "benchmark_tpch_orders")
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return seedTPCH(ctx, db, scaleFactor)
}

func seedTPCH(ctx context.Context, db *sql.DB, scaleFactor int) error {
	if scaleFactor < 1 {
		scaleFactor = 1
	}
	customers := 1000 * scaleFactor
	parts := 1000 * scaleFactor
	orders := 3000 * scaleFactor

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i := 0; i < 5; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_tpch_region(r_regionkey, r_name) VALUES (?, ?)`, i, fmt.Sprintf("REGION%d", i)); err != nil {
			return err
		}
	}
	for i := 1; i <= 25; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_tpch_nation(n_nationkey, n_name, n_regionkey) VALUES (?, ?, ?)`, i, fmt.Sprintf("NATION%d", i), i%5); err != nil {
			return err
		}
	}
	for i := 1; i <= 200*scaleFactor; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_tpch_supplier(s_suppkey, s_name, s_nationkey, s_acctbal) VALUES (?, ?, ?, ?)`, i, fmt.Sprintf("Supplier %04d", i), (i%25)+1, float64(i%10000)); err != nil {
			return err
		}
	}
	segments := []string{"AUTOMOBILE", "BUILDING", "FURNITURE", "MACHINERY", "HOUSEHOLD"}
	for i := 1; i <= customers; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_tpch_customer(c_custkey, c_name, c_nationkey, c_acctbal, c_mktsegment) VALUES (?, ?, ?, ?, ?)`, i, fmt.Sprintf("Customer %04d", i), (i%25)+1, float64(i%9000), segments[i%len(segments)]); err != nil {
			return err
		}
	}
	for i := 1; i <= parts; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_tpch_part(p_partkey, p_name, p_mfgr, p_brand, p_type, p_size, p_retailprice) VALUES (?, ?, ?, ?, ?, ?, ?)`, i, fmt.Sprintf("Part %04d", i), fmt.Sprintf("MFGR%d", i%5), fmt.Sprintf("Brand#%d", i%20), fmt.Sprintf("TYPE%d", i%8), (i%50)+1, float64(100+(i%1000))); err != nil {
			return err
		}
		for s := 1; s <= 4; s++ {
			supp := ((i + s) % (200 * scaleFactor)) + 1
			if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_tpch_partsupp(ps_partkey, ps_suppkey, ps_availqty, ps_supplycost) VALUES (?, ?, ?, ?)`, i, supp, (i*s)%9999, float64((i+s)%500)+1); err != nil {
				return err
			}
		}
	}
	priorities := []string{"1-URGENT", "2-HIGH", "3-MEDIUM", "4-NOT SPECIFIED", "5-LOW"}
	modes := []string{"AIR", "RAIL", "TRUCK", "MAIL", "SHIP"}
	for i := 1; i <= orders; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_tpch_orders(o_orderkey, o_custkey, o_orderstatus, o_totalprice, o_orderdate, o_orderpriority) VALUES (?, ?, ?, ?, DATE_ADD('2025-01-01', INTERVAL ? DAY), ?)`, i, (i%customers)+1, string("FOP"[i%3]), float64(100+(i%10000)), i%730, priorities[i%len(priorities)]); err != nil {
			return err
		}
		for line := 1; line <= 4; line++ {
			part := ((i * line) % parts) + 1
			supp := ((part + line) % (200 * scaleFactor)) + 1
			if _, err := tx.ExecContext(ctx, `INSERT INTO benchmark_tpch_lineitem(l_orderkey, l_partkey, l_suppkey, l_linenumber, l_quantity, l_extendedprice, l_discount, l_tax, l_returnflag, l_linestatus, l_shipdate, l_commitdate, l_receiptdate, l_shipmode) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, DATE_ADD('2025-01-02', INTERVAL ? DAY), DATE_ADD('2025-01-04', INTERVAL ? DAY), DATE_ADD('2025-01-06', INTERVAL ? DAY), ?)`, i, part, supp, line, float64((i+line)%50)+1, float64((i+line)%5000)+10, float64((i+line)%10)/100, float64((i+line)%8)/100, string("RAN"[i%3]), string("OF"[i%2]), i%730, i%730, i%730, modes[i%len(modes)]); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func TPCHCases(db *sql.DB) []benchmark.Case {
	q := func(sql string) func(context.Context) error {
		return func(ctx context.Context) error { return runQuery(ctx, db, sql) }
	}
	return []benchmark.Case{
		{Name: "Q01", Description: "Pricing summary", Run: q(`SELECT l_returnflag, l_linestatus, SUM(l_quantity), SUM(l_extendedprice), AVG(l_discount), COUNT(*) FROM benchmark_tpch_lineitem WHERE l_shipdate <= '2026-12-01' GROUP BY l_returnflag, l_linestatus`)},
		{Name: "Q02", Description: "Minimum supply cost", Run: q(`SELECT p.p_partkey, s.s_suppkey, ps.ps_supplycost FROM benchmark_tpch_part p JOIN benchmark_tpch_partsupp ps ON ps.ps_partkey = p.p_partkey JOIN benchmark_tpch_supplier s ON s.s_suppkey = ps.ps_suppkey WHERE p.p_size BETWEEN 10 AND 20 ORDER BY ps.ps_supplycost LIMIT 100`)},
		{Name: "Q03", Description: "Shipping priority", Run: q(`SELECT o.o_orderkey, SUM(l.l_extendedprice * (1-l.l_discount)) revenue FROM benchmark_tpch_customer c JOIN benchmark_tpch_orders o ON o.o_custkey = c.c_custkey JOIN benchmark_tpch_lineitem l ON l.l_orderkey = o.o_orderkey WHERE c.c_mktsegment='BUILDING' AND o.o_orderdate < '2025-08-01' GROUP BY o.o_orderkey ORDER BY revenue DESC LIMIT 100`)},
		{Name: "Q04", Description: "Order priority count", Run: q(`SELECT o_orderpriority, COUNT(*) FROM benchmark_tpch_orders WHERE o_orderdate >= '2025-01-01' AND o_orderdate < '2025-04-01' AND EXISTS (SELECT 1 FROM benchmark_tpch_lineitem WHERE l_orderkey=o_orderkey AND l_commitdate < l_receiptdate) GROUP BY o_orderpriority`)},
		{Name: "Q05", Description: "Local supplier volume", Run: q(`SELECT n.n_name, SUM(l.l_extendedprice*(1-l.l_discount)) revenue FROM benchmark_tpch_customer c JOIN benchmark_tpch_orders o ON o.o_custkey=c.c_custkey JOIN benchmark_tpch_lineitem l ON l.l_orderkey=o.o_orderkey JOIN benchmark_tpch_supplier s ON s.s_suppkey=l.l_suppkey JOIN benchmark_tpch_nation n ON n.n_nationkey=s.s_nationkey JOIN benchmark_tpch_region r ON r.r_regionkey=n.n_regionkey WHERE r.r_name='REGION1' GROUP BY n.n_name ORDER BY revenue DESC`)},
		{Name: "Q06", Description: "Forecasting revenue", Run: q(`SELECT SUM(l_extendedprice*l_discount) FROM benchmark_tpch_lineitem WHERE l_shipdate >= '2025-01-01' AND l_shipdate < '2026-01-01' AND l_discount BETWEEN 0.03 AND 0.07 AND l_quantity < 24`)},
		{Name: "Q07", Description: "Volume shipping", Run: q(`SELECT n1.n_name, n2.n_name, YEAR(l.l_shipdate), SUM(l.l_extendedprice*(1-l.l_discount)) FROM benchmark_tpch_supplier s JOIN benchmark_tpch_lineitem l ON l.l_suppkey=s.s_suppkey JOIN benchmark_tpch_orders o ON o.o_orderkey=l.l_orderkey JOIN benchmark_tpch_customer c ON c.c_custkey=o.o_custkey JOIN benchmark_tpch_nation n1 ON n1.n_nationkey=s.s_nationkey JOIN benchmark_tpch_nation n2 ON n2.n_nationkey=c.c_nationkey GROUP BY n1.n_name, n2.n_name, YEAR(l.l_shipdate) LIMIT 100`)},
		{Name: "Q08", Description: "Market share", Run: q(`SELECT YEAR(o.o_orderdate), SUM(CASE WHEN n.n_name='NATION1' THEN l.l_extendedprice*(1-l.l_discount) ELSE 0 END) / SUM(l.l_extendedprice*(1-l.l_discount)) FROM benchmark_tpch_part p JOIN benchmark_tpch_lineitem l ON l.l_partkey=p.p_partkey JOIN benchmark_tpch_supplier s ON s.s_suppkey=l.l_suppkey JOIN benchmark_tpch_nation n ON n.n_nationkey=s.s_nationkey JOIN benchmark_tpch_orders o ON o.o_orderkey=l.l_orderkey WHERE p.p_type='TYPE1' GROUP BY YEAR(o.o_orderdate)`)},
		{Name: "Q09", Description: "Product profit", Run: q(`SELECT n.n_name, YEAR(o.o_orderdate), SUM(l.l_extendedprice*(1-l.l_discount)-ps.ps_supplycost*l.l_quantity) profit FROM benchmark_tpch_part p JOIN benchmark_tpch_lineitem l ON l.l_partkey=p.p_partkey JOIN benchmark_tpch_partsupp ps ON ps.ps_partkey=l.l_partkey AND ps.ps_suppkey=l.l_suppkey JOIN benchmark_tpch_orders o ON o.o_orderkey=l.l_orderkey JOIN benchmark_tpch_supplier s ON s.s_suppkey=l.l_suppkey JOIN benchmark_tpch_nation n ON n.n_nationkey=s.s_nationkey WHERE p.p_name LIKE 'Part%' GROUP BY n.n_name, YEAR(o.o_orderdate) LIMIT 100`)},
		{Name: "Q10", Description: "Returned item reporting", Run: q(`SELECT c.c_custkey, c.c_name, SUM(l.l_extendedprice*(1-l.l_discount)) revenue FROM benchmark_tpch_customer c JOIN benchmark_tpch_orders o ON o.o_custkey=c.c_custkey JOIN benchmark_tpch_lineitem l ON l.l_orderkey=o.o_orderkey WHERE o.o_orderdate >= '2025-01-01' AND l.l_returnflag='R' GROUP BY c.c_custkey, c.c_name ORDER BY revenue DESC LIMIT 100`)},
		{Name: "Q11", Description: "Important stock", Run: q(`SELECT ps_partkey, SUM(ps_supplycost*ps_availqty) value FROM benchmark_tpch_partsupp JOIN benchmark_tpch_supplier ON s_suppkey=ps_suppkey WHERE s_nationkey=1 GROUP BY ps_partkey HAVING value > 1000 ORDER BY value DESC LIMIT 100`)},
		{Name: "Q12", Description: "Shipping modes", Run: q(`SELECT l_shipmode, SUM(CASE WHEN o_orderpriority IN ('1-URGENT','2-HIGH') THEN 1 ELSE 0 END), SUM(CASE WHEN o_orderpriority NOT IN ('1-URGENT','2-HIGH') THEN 1 ELSE 0 END) FROM benchmark_tpch_orders JOIN benchmark_tpch_lineitem ON l_orderkey=o_orderkey WHERE l_shipmode IN ('MAIL','SHIP') GROUP BY l_shipmode`)},
		{Name: "Q13", Description: "Customer distribution", Run: q(`SELECT c_count, COUNT(*) FROM (SELECT c.c_custkey, COUNT(o.o_orderkey) c_count FROM benchmark_tpch_customer c LEFT JOIN benchmark_tpch_orders o ON o.o_custkey=c.c_custkey GROUP BY c.c_custkey) x GROUP BY c_count ORDER BY c_count DESC`)},
		{Name: "Q14", Description: "Promotion effect", Run: q(`SELECT 100 * SUM(CASE WHEN p.p_type='TYPE1' THEN l.l_extendedprice*(1-l.l_discount) ELSE 0 END) / SUM(l.l_extendedprice*(1-l.l_discount)) FROM benchmark_tpch_lineitem l JOIN benchmark_tpch_part p ON p.p_partkey=l.l_partkey WHERE l.l_shipdate >= '2025-03-01' AND l.l_shipdate < '2025-04-01'`)},
		{Name: "Q15", Description: "Top supplier", Run: q(`SELECT s.s_suppkey, s.s_name, SUM(l.l_extendedprice*(1-l.l_discount)) revenue FROM benchmark_tpch_supplier s JOIN benchmark_tpch_lineitem l ON l.l_suppkey=s.s_suppkey GROUP BY s.s_suppkey, s.s_name ORDER BY revenue DESC LIMIT 10`)},
		{Name: "Q16", Description: "Parts supplier relationship", Run: q(`SELECT p_brand, p_type, p_size, COUNT(DISTINCT ps_suppkey) FROM benchmark_tpch_partsupp JOIN benchmark_tpch_part ON p_partkey=ps_partkey WHERE p_brand <> 'Brand#1' GROUP BY p_brand, p_type, p_size LIMIT 100`)},
		{Name: "Q17", Description: "Small quantity revenue", Run: q(`SELECT SUM(l_extendedprice)/7 FROM benchmark_tpch_lineitem l JOIN benchmark_tpch_part p ON p.p_partkey=l.l_partkey WHERE p.p_brand='Brand#1' AND l.l_quantity < (SELECT 0.2*AVG(l2.l_quantity) FROM benchmark_tpch_lineitem l2 WHERE l2.l_partkey=p.p_partkey)`)},
		{Name: "Q18", Description: "Large volume customer", Run: q(`SELECT c.c_name, c.c_custkey, o.o_orderkey, SUM(l.l_quantity) FROM benchmark_tpch_customer c JOIN benchmark_tpch_orders o ON o.o_custkey=c.c_custkey JOIN benchmark_tpch_lineitem l ON l.l_orderkey=o.o_orderkey GROUP BY c.c_name, c.c_custkey, o.o_orderkey HAVING SUM(l.l_quantity) > 120 ORDER BY SUM(l.l_quantity) DESC LIMIT 100`)},
		{Name: "Q19", Description: "Discounted revenue", Run: q(`SELECT SUM(l.l_extendedprice*(1-l.l_discount)) FROM benchmark_tpch_lineitem l JOIN benchmark_tpch_part p ON p.p_partkey=l.l_partkey WHERE p.p_brand IN ('Brand#1','Brand#2','Brand#3') AND l.l_quantity BETWEEN 1 AND 30`)},
		{Name: "Q20", Description: "Potential promotion", Run: q(`SELECT s.s_name FROM benchmark_tpch_supplier s WHERE s.s_suppkey IN (SELECT ps.ps_suppkey FROM benchmark_tpch_partsupp ps WHERE ps.ps_partkey IN (SELECT p.p_partkey FROM benchmark_tpch_part p WHERE p.p_name LIKE 'Part 00%')) ORDER BY s.s_name LIMIT 100`)},
		{Name: "Q21", Description: "Supplier wait", Run: q(`SELECT s.s_name, COUNT(*) numwait FROM benchmark_tpch_supplier s JOIN benchmark_tpch_lineitem l1 ON s.s_suppkey=l1.l_suppkey JOIN benchmark_tpch_orders o ON o.o_orderkey=l1.l_orderkey WHERE o.o_orderstatus='F' AND l1.l_receiptdate > l1.l_commitdate GROUP BY s.s_name ORDER BY numwait DESC LIMIT 100`)},
		{Name: "Q22", Description: "Global sales opportunity", Run: q(`SELECT SUBSTRING(c_name, 1, 2) country_code, COUNT(*), SUM(c_acctbal) FROM benchmark_tpch_customer WHERE c_acctbal > (SELECT AVG(c_acctbal) FROM benchmark_tpch_customer WHERE c_acctbal > 0) GROUP BY SUBSTRING(c_name, 1, 2)`)},
	}
}
