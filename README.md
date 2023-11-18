# flash-card-manager

## Описание проекта
Данный проект создан для управления флеш-картами (часто называемыми "анки картами") для интервального повторения.
Основными объектами в нашем приложении являются колоды и карты. 
Каждая карта принадлежит определенной колоде и содержит информацию на лицевой и обратной сторонах.
Разработано с поддержкой протоколов HTTP и gRPC. Интеграция с Kafka улучшает возможности потоковой передачи данных в реальном времени. Внедрено структурное логирование через Zap и трассировка через Jaeger

## Стек технологий
- Go
- gRPC + RESTful gateway
- Postgres
- Goose
- Kafka
- Docker
- Zap (uber)
- Gomock
- Testify
- Jaeger

## Тестирование
- Написаны unit-тесты для хэндлеров, репозитория 
- Написаны интеграционные тесты для Базы данных и Kafka

## Пользовательская инструкция

Прежде чем запустить тесты, убедитесь, что у вас установлен `goose`:

```bash
go get -u github.com/pressly/goose/cmd/goose
```

### Миграции через Goose

**Создание миграции**
```make migration-create name=имя_вашей_миграции```

**Применение миграций**
```make test-migration-up```

**Откат миграций**
```make test-migration-down```

### Тестовое окружение

**Запуск тестового окружения (Docker Compose)**
```make test-env-up```

**Остановка тестового окружения (Docker Compose)**
```make test-env-down```

**Очистка данных в базе данных тестового окружения**
```make clean-db```

### Запуск тестов

**Запуск интеграционных тестов**
```make integration-tests```

**Запуск юнит-тестов**
```make unit-tests```

**Запуск всех тестов (миграции, юнит-тесты, интеграционные тесты)**
```make test-all```

`Убедитесь, что перед использованием тестового окружения и миграций переменная POSTGRES_SETUP_TEST настроена правильно в вашем Makefile.`

```
ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := user=test password=test dbname=test host=localhost port=5432 sslmode=disable
endif
```

# Взаимодествие с сервисом

## Колоды

**Создание колоды**
```go run cmd/client/main.go -addr=localhost:9000 createDeck <Deck titile> <Description of deck> <Author>```

**Получение колоды по ID**
```go run cmd/client/main.go -addr=localhost:9000 getDeckById <Deck ID>```

**Обновление колоды**
```go run cmd/client/main.go -addr=localhost:9000 updateDeck <Deck ID> <Updated Title> <Updated Description> <Updated Author>```

**Удаление колоды по ID**
```go run cmd/client/main.go -addr=localhost:9000 deleteDeck <Deck ID>```

## Карты

**Создание карты**
```go run cmd/client/main.go -addr=localhost:9000 createCard  <Front card> <Back card> <Deck ID> <Author>```

**Получение карты по ID**
```go run cmd/client/main.go -addr=localhost:9000 getCardById <Card ID>```

**Обновление карты**
```go run cmd/client/main.go -addr=localhost:9000 updateCard <Card ID> <Updated Front card> <Updated Back card> <Deck ID> <Updated Author>```

**Удаление карты по ID**
```go run cmd/client/main.go -addr=localhost:9000 deleteCard <Card ID>```





