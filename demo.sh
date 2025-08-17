#!/bin/bash
echo "❗ Демонстрация сервиса обработки заказов из Kafka с сохранением в PostgreSQL и кэшированием..."

echo "📦 Запуск сервисов через Docker Compose..."
make up
sleep 3

echo "🏥 Проверка здоровья..."
curl http://localhost:8081/health | jq .
sleep 3

echo "🔍 Доказательство того, что база данных пустая (test-1 нет)..."
curl http://localhost:8081/order/test-1 | jq .
sleep 3

echo "🔍 Доказательство того, что база данных пустая (b563feb7b2b84b6test нет)..."
curl http://localhost:8081/order/b563feb7b2b84b6test | jq .
sleep 3

echo "📨 Отправка тестовых данных в Kafka (эмуляция продюсера)..."
make test-kafka
sleep 2

echo "🔍 Логи, после отправки тестовых данных в Kafka..."
docker compose logs app --tail=18
sleep 10

echo "🌐 Тестируем API..."
curl http://localhost:8081/order/test-1 | jq .
sleep 7

echo "🔄 Рестарт сервиса..."
docker compose restart app
sleep 3

echo "✅ Проверяем восстановление кэша..."
docker compose logs app --tail=5
sleep 5

echo "💾 ДЕМОНСТРАЦИЯ КЭША - останавливаем БД..."
docker compose stop db
sleep 3

echo "📦 Получаем заказ test-1 при недоступной БД (из кэша)..."
curl http://localhost:8081/order/test-1 | jq .
sleep 5

echo "🔍 Смотрим логи после получения test-1..."
docker compose logs app --tail=5
sleep 7

echo "📦 Получаем заказ b563feb7b2b84b6test при недоступной БД (из кэша)..."
curl http://localhost:8081/order/b563feb7b2b84b6test | jq .
sleep 5

echo "🔍 Смотрим логи после получения b563feb7b2b84b6test..."
docker compose logs app --tail=5
sleep 7

echo "🔧 Запускаем БД обратно..."
docker compose start db
sleep 10

echo "✅ Демонстрация сервиса завершена!"
echo "❗ Заказы: test-1 и b563feb7b2b84b6test (можно скопировать для веб-интерфейса)"
echo "🌐 Веб-интерфейс: http://localhost:8080"