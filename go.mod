module ride-together-bot

go 1.21

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/pressly/goose/v3 v3.20.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/sethvargo/go-retry v0.2.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
)

replace github.com/go-telegram-bot-api/telegram-bot-api/v5 => github.com/iamwavecut/telegram-bot-api v0.0.0-20240102154128-a33af0365ce6
