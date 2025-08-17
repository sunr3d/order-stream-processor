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
cat data/test_1.json | tr -d '\n' | docker run --rm -i --network order-stream-processor_default \
  bitnami/kafka:latest kafka-console-producer.sh \
  --bootstrap-server kafka:9092 \
  --topic orders

# Ждем обработки второго заказа
echo "⏳ Ждем обработки второго заказа..."
sleep 5

# Получаем второй заказ
echo "📥 Получаем второй заказ..."
curl -s "http://localhost:8081/order/test-1" | jq .

# Рестарт сервиса
echo "Производим рестарт сервиса..."
docker compose restart app

# Ждем рестарта
echo "⏳ Ждем рестарта..."
sleep 5

# Получаем заказ 1
echo "📥 Получаем заказ 1..."
curl -s "http://localhost:8081/order/b563feb7b2b84b6test" | jq .

# Получаем заказ 2
echo "📥 Получаем заказ 2..."
curl -s "http://localhost:8081/order/test-1" | jq .

echo "✅ Тестирование завершено!"