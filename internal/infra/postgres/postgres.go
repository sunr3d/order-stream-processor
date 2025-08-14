package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/config"
	"github.com/sunr3d/order-stream-processor/internal/interfaces/infra"
	"github.com/sunr3d/order-stream-processor/models"
)

const (
	queryCreate  = `INSERT INTO orders (order_uid, data) VALUES ($1, $2)`
	queryRead    = `SELECT data FROM orders WHERE order_uid = $1`
	queryReadAll = `SELECT data FROM orders ORDER BY order_uid`
)

var _ infra.Database = (*postgresRepo)(nil)

type postgresRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func New(cfg config.PostgresConfig, log *zap.Logger) (infra.Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.PingContext: %w", err)
	}

	log.Info("соединение с PostgreSQL установлено")

	return &postgresRepo{
		db:     db,
		logger: log,
	}, nil
}

func (r *postgresRepo) Close() error {
	return r.db.Close()
}

func (r *postgresRepo) Create(ctx context.Context, order *models.Order) error {
	logger := r.logger.With(
		zap.String("op", "postgres.Create"),
		zap.String("order_uid", order.OrderUID),
	)

	logger.Info("сохранение заказа в БД...")

	data, err := json.Marshal(order)
	if err != nil {
		logger.Error("ошибка при маршалинге заказа", zap.Error(err))
		return fmt.Errorf("json.Marshal: %w", err)
	}

	_, err = r.db.ExecContext(ctx, queryCreate, order.OrderUID, data)
	if err != nil {
		logger.Error("ошибка при сохранении заказа в БД", zap.Error(err))
		return fmt.Errorf("db.ExecContext: %w", err)
	}

	logger.Info("заказ успешно сохранен в БД")
	return nil
}

func (r *postgresRepo) Read(ctx context.Context, orderUID string) (*models.Order, error) {
	logger := r.logger.With(
		zap.String("op", "postgres.Read"),
		zap.String("order_uid", orderUID),
	)

	logger.Info("поиск заказа в БД...")

	var data []byte
	err := r.db.QueryRowContext(ctx, queryRead, orderUID).Scan(&data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("заказ не найден")
			return nil, fmt.Errorf("заказ не найден: %s", orderUID)
		}
		logger.Error("ошибка чтения из БД", zap.Error(err))
		return nil, fmt.Errorf("db.QueryRowContext: %w", err)
	}

	var order models.Order
	if err := json.Unmarshal(data, &order); err != nil {
		logger.Error("ошибка при парсинге заказа", zap.Error(err))
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	logger.Info("заказ успешно найден в БД")
	return &order, nil
}

func (r *postgresRepo) ReadAll(ctx context.Context) ([]*models.Order, error) {
	logger := r.logger.With(
		zap.String("op", "postgres.ReadAll"),
	)

	logger.Info("получение всех заказов из БД...")

	rows, err := r.db.QueryContext(ctx, queryReadAll)
	if err != nil {
		logger.Error("ошибка при получении всех заказов из БД", zap.Error(err))
		return nil, fmt.Errorf("db.QueryContext: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			logger.Error("ошибка при записи строки из БД", zap.Error(err))
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		var order models.Order
		if err := json.Unmarshal(data, &order); err != nil {
			logger.Error("ошибка при парсинге заказа", zap.Error(err))
			return nil, fmt.Errorf("json.Unmarshal: %w", err)
		}

		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		logger.Error("произошла ошибка во время чтения строк из БД", zap.Error(err))
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	logger.Info("все заказы успешно получены из БД", zap.Int("count", len(orders)))
	return orders, nil
}
