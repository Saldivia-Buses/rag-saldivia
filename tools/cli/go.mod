module github.com/Camionerou/rag-saldivia/tools/cli

go 1.25.0

require (
	github.com/Camionerou/rag-saldivia/tools/pkg v0.0.0
	github.com/go-sql-driver/mysql v1.9.3
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.9.1
	github.com/shopspring/decimal v1.4.0
	github.com/spf13/cobra v1.10.2
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace github.com/Camionerou/rag-saldivia/tools/pkg => ../pkg
