#!/bin/bash

cat data/model.json | tr -d '\n' | docker run --rm -i --network order-stream-processor_default \
  bitnami/kafka:latest kafka-console-producer.sh \
  --bootstrap-server kafka:9092 \
  --topic orders

cat data/test_1.json | tr -d '\n' | docker run --rm -i --network order-stream-processor_default \
  bitnami/kafka:latest kafka-console-producer.sh \
  --bootstrap-server kafka:9092 \
  --topic orders