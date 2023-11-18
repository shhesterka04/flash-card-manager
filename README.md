# flash-card-manager

Данный проект создан для управления флеш-картами 
(часто называемыми "анки картами") для интервального повторения.
Основными объектами в нашем приложении являются колоды и карты. 
Каждая карта принадлежит определенной колоде и содержит информацию на лицевой и обратной сторонах.
Разработано с поддержкой протоколов HTTP и gRPC. Интеграция с Kafka улучшает возможности потоковой передачи данных в реальном времени.


# БД в Docker

**Запуск контейнера**
```docker-compose up -d```


# Миграции через Goose

**Создание миграции**
```make migration-create name=имя_вашей_миграции```

**Применение миграций**
```make test-migration-up```

**Откат миграций**
```make test-migration-down```

# Curl
## Колоды

**Создание колоды**
```go run cmd/client/main.go -addr=localhost:9000 createDeck "My Deck Title" "Description of my deck" "Author Name"```

**Получение колоды по ID**
```go run cmd/client/main.go -addr=localhost:9000 getDeckById 1```

**Обновление колоды**
```go run cmd/client/main.go -addr=localhost:9000 updateDeck 1 "Updated Title" "Updated Description" "Author Name"```

**Удаление колоды по ID**
```go run cmd/client/main.go -addr=localhost:9000 deleteDeck 1```

## Карты

**Создание карты**
```go run cmd/client/main.go -addr=localhost:9000 createCard  "Front card" "Back card" 2  "John Doe"```

**Получение карты по ID**
```go run cmd/client/main.go -addr=localhost:9000 getCardById 3```

**Обновление карты**
```go run cmd/client/main.go -addr=localhost:9000 updateCard 6 "Updated Front card" "Updated Back card" 2 "John Doe"```


**Удаление карты по ID**
```go run cmd/client/main.go -addr=localhost:9000 deleteCard 6```





