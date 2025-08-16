#!/bin/bash

echo "Тестирование Kafka консьюмера..."

# Проверяем health
echo "🏥 Health check..."
curl -s http://localhost:8081/health | jq .

# Отправляем в Kafka
echo "Отправляем в Kafka заказ из тестового файла..."
cat data/model.json | tr -d '\n' | docker run --rm -i --network order-stream-processor_default \
  bitnami/kafka:latest kafka-console-producer.sh \
  --bootstrap-server kafka:9092 \
  --topic orders

# Ждем обработки
echo "⏳ Ждем обработки..."
sleep 5

# Получаем заказ
echo "📥 Получаем заказ..."
curl -s "http://localhost:8081/order/b563feb7b2b84b6test" | jq .

# Отправляем в Kafka еще один заказ
echo "Отправляем в Kafka еще один заказ..."
echo '{"order_uid": "test-2", "customer_id": "customer-2", "track_number": "TRACK-002", "items": [{"name": "Test Item 2", "price": 200, "total_price": 200}], "delivery": {"name": "User 2"}, "payment": {"transaction": "tx-2", "provider": "test", "amount": 200, "payment_dt": 1640995200}}' | docker run --rm -i --network order-stream-processor_default \
  bitnami/kafka:latest kafka-console-producer.sh \
  --bootstrap-server kafka:9092 \
  --topic orders

# Ждем обработки второго заказа
echo "⏳ Ждем обработки второго заказа..."
sleep 5

# Получаем второй заказ
echo "📥 Получаем второй заказ..."
curl -s "http://localhost:8081/order/test-2" | jq .

# Рестарт сервиса
echo "Производим рестарт сервиса..."
docker compose restart app

# Ждем рестарта
echo "⏳ Ждем рестарта..."
sleep 10

# Получаем заказ
echo "📥 Получаем заказ 1..."
curl -s "http://localhost:8081/order/b563feb7b2b84b6test" | jq .

# Получаем все заказы
echo "📥 Получаем заказ 2..."
curl -s "http://localhost:8081/order/test-2" | jq .

echo "✅ Тестирование завершено!"