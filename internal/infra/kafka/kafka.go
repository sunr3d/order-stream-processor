package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/config"
	"github.com/sunr3d/order-stream-processor/internal/interfaces/infra"
)

var _ infra.Broker = (*kafkaBroker)(nil)

type kafkaBroker struct {
	client    sarama.Client
	consumers sarama.ConsumerGroup
	handler   func(context.Context, []byte) error
	config    config.KafkaConfig
	logger    *zap.Logger
}

func New(cfg config.KafkaConfig, logger *zap.Logger) (infra.Broker, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewClient(cfg.Brokers, config)
	if err != nil {
		logger.Error("ошибка подключения к Kafka", zap.Error(err))
		return nil, fmt.Errorf("не удалось подключиться к Kafka: %w", err)
	}

	consumers, err := sarama.NewConsumerGroupFromClient(cfg.GroupID, client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("не удалось создать consumer group: %w", err)
	}

	return &kafkaBroker{
		client:    client,
		consumers: consumers,
		config:    cfg,
		logger:    logger,
	}, nil
}

func (b *kafkaBroker) StartConsumer(ctx context.Context) error {
	logger := b.logger.With(
		zap.String("op", "kafka.Start"),
	)

	logger.Info("запуск Kafka consumers group",
		zap.String("group_id", b.config.GroupID),
		zap.String("topic", b.config.Topic),
		zap.String("brokers", strings.Join(b.config.Brokers, ", ")),
	)

	for {
		err := b.consumers.Consume(ctx, []string{b.config.Topic}, b)
		if err != nil {
			logger.Error("ошибка при чтении сообщений из Kafka", zap.Error(err))
		}

		if ctx.Err() != nil {
			logger.Info("остановка Kafka consumers group по причине контекста",
				zap.String("group_id", b.config.GroupID),
				zap.String("topic", b.config.Topic),
				zap.String("brokers", strings.Join(b.config.Brokers, ", ")),
			)
			break
		}
	}

	return nil
}

func (b *kafkaBroker) Stop() error {
	logger := b.logger.With(
		zap.String("op", "kafka.Stop"),
	)

	logger.Info("закрытие соединения с Kafka",
		zap.String("group_id", b.config.GroupID),
		zap.String("topic", b.config.Topic),
		zap.String("brokers", strings.Join(b.config.Brokers, ", ")),
	)

	b.consumers.Close()
	return b.client.Close()
}

func (b *kafkaBroker) SetHandler(handler func(context.Context, []byte) error) {
	b.handler = handler
}

func (b *kafkaBroker) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger := b.logger.With(
		zap.String("op", "kafka.ConsumeClaim"),
	)

	for msg := range claim.Messages() {
		logger.Info("получено сообщение из Kafka",
			zap.Int32("partition", msg.Partition),
			zap.Int64("offset", msg.Offset),
			zap.String("key (order_uid)", string(msg.Key)),
		)

		for attempt := 1; attempt <= b.config.MaxRetries; attempt++ {
			if err := b.handler(session.Context(), msg.Value); err != nil {
				logger.Error("ошибка при обработке сообщения",
					zap.Int32("partition", msg.Partition),
					zap.Int64("offset", msg.Offset),
					zap.String("key (order_uid)", string(msg.Key)),
					zap.Int("attempt", attempt),
					zap.Int("max_retries", b.config.MaxRetries),
					zap.Error(err),
				)

				if attempt == b.config.MaxRetries {
					session.MarkMessage(msg, "")
					logger.Warn("превышено количество попыток обработки сообщения, сообщение будет пропущено",
						zap.Int32("partition", msg.Partition),
						zap.Int64("offset", msg.Offset),
						zap.String("key (order_uid)", string(msg.Key)),
					)
				}

				continue
			} else {
				break
			}
		}

		session.MarkMessage(msg, "")
		logger.Info("сообщение обработано успешно",
			zap.Int32("partition", msg.Partition),
			zap.Int64("offset", msg.Offset),
			zap.String("key (order_uid)", string(msg.Key)),
		)
	}

	return nil
}

func (b *kafkaBroker) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (b *kafkaBroker) Cleanup(sarama.ConsumerGroupSession) error { return nil }
