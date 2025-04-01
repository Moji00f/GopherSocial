include .envrc

MIGRATION_PATH=./cmd/migrate/migrations

.PHONY: migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATION_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_MIGRATOR_ADDR) up

.PHONY: migrate-status
migrate-status:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_MIGRATOR_ADDR) version

#جای x آخرین نسخه سالم مایگریشن را قرار بده
# force x
.PHONY: clear-dirty
clear-dirty:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_MIGRATOR_ADDR) force 11

#make migrate-down 3 (3 is step of number of rollback)
.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_MIGRATOR_ADDR) down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: seed
seed:
	@go run cmd/migrate/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt
