# Third-party Licenses and Dataset Notes

## Go MySQL Driver

This project uses `github.com/go-sql-driver/mysql`, licensed under the Mozilla Public License 2.0.

Source: <https://github.com/go-sql-driver/mysql>

## Sakila

This repository does not redistribute Oracle's Sakila SQL files. The CLI creates benchmark-owned Sakila-style tables and synthetic data using names prefixed with `benchmark_sakila_`.

Oracle documents the Sakila sample database separately. The Sakila schema and data files are described by Oracle as available under a New BSD license, while the accompanying documentation is subject to Oracle's documentation terms.

Source: <https://dev.mysql.com/doc/sakila/en/>

## TPC-H

This repository does not redistribute TPC-H tools, dbgen output, or official TPC-H assets. The CLI includes a TPC-H-inspired workload for MySQL using benchmark-owned tables prefixed with `benchmark_tpch_`.

The output intentionally uses wording such as "TPC-H-inspired workload" and does not claim official TPC-H compliance.

Source: <https://www.tpc.org/tpc_documents_current_versions/current_specifications5.asp>
