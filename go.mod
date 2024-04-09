module ride-together-bot

go 1.21

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
)

require filippo.io/edwards25519 v1.1.0 // indirect

replace github.com/go-telegram-bot-api/telegram-bot-api/v5 => github.com/iamwavecut/telegram-bot-api v0.0.0-20240102154128-a33af0365ce6
