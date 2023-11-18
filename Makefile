ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := user=test password=test dbname=test host=localhost port=5432 sslmode=disable
endif

MIGRATION_FOLDER=$(CURDIR)/scripts/migrations
DOCKER_COMPOSE_FILE=docker-compose.yaml

.PHONY: migration-create test-migration-up test-migration-down
.PHONY: test-env-up test-env-down integration-tests unit-tests clean-db

migration-create:
	goose -dir "$(MIGRATION_FOLDER)" create "$(name)" sql

test-migration-up:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" up

test-migration-down:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" down

test-env-up:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

test-env-down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

integration-tests: test-env-up
	go test ./... -tags=integration
	make test-env-down


unit-tests:
	go test ./... -tags=unit


clean-db: test-env-up
	docker-compose -f docker-compose.yaml exec db psql "user=test password=test dbname=test sslmode=disable" -c "DELETE FROM decks;"

test-all:
	make test-env-up
	make test-migration-up
	make unit-tests
	make integration-tests