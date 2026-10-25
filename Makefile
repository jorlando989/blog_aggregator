DB_URL=postgres://postgres:postgres@localhost:5432/gator

reset:
	goose -dir sql/schema postgres "$(DB_URL)" down-to 0
	goose -dir sql/schema postgres "$(DB_URL)" up