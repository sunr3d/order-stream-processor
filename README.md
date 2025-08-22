# Order Stream Processor

Демонстрационный сервис для обработки заказов из Kafka с сохранением в PostgreSQL и кэшированием.


## Быстрый старт

### Запустить демо (требуется Docker)
```bash
make demo
```

## Конфигурация

```bash
HTTP_PORT=8081
LOG_LEVEL=info
POSTGRES_HOST=db
POSTGRES_PORT=5432
POSTGRES_USER=orders_user
POSTGRES_PASSWORD=orders_password
POSTGRES_DB=orders
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC=orders
KAFKA_GROUP_ID=order-processor
```

## API

### Создание заказа

```bash
curl -X POST http://localhost:8081/order \
  -H "Content-Type: application/json" \
  -d @data/model.json
```

### Получение заказа
```bash
curl http://localhost:8081/order/b563feb7b2b84b6test
```

### Проверка работоспособности сервиса
```bash
curl http://localhost:8081/health
```

## Kafka

### Создание заказа
```bash
cat data/model.json | tr -d '\n' | docker run --rm -i --network order-stream-processor_default \
  bitnami/kafka:latest kafka-console-producer.sh \
  --bootstrap-server kafka:9092 \
  --topic orders
```

## Веб-интерфейс

http://localhost:8080 для поиска заказов через веб-интерфейс.

## Структура проекта

- `models/` - доменные модели сервиса
- `internal/services/` - бизнес-логика сервиса обработки заказов
- `internal/infra/` - PostgreSQL, Kafka, in-memory кэш
- `internal/handlers/` - HTTP и Kafka обработчики
- `internal/server/` - HTTP сервер с graceful shutdown
- `internal/interfaces/` - инфраструктурные и сервисные интерфейсы

## Команды

```bash
make demo        # Запуск демонстрации сервиса
make up          # Запуск сервисов
make down        # Остановка сервисов
make clean       # Остановка сервисов с очисткой томов
make test        # Запуск юни-тестов
make test-kafka  # Скрипт-эмулятор продюсера кафки (отправляет два заказа)
```
