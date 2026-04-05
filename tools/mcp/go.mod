module github.com/Camionerou/rag-saldivia/tools/mcp

go 1.25.0

require github.com/Camionerou/rag-saldivia/tools/pkg v0.0.0

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.9.1 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace github.com/Camionerou/rag-saldivia/tools/pkg => ../pkg
