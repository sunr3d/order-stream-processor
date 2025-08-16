package infra

import "context"

//go:generate go run github.com/vektra/mockery/v2@v2.53.2 --name=Broker --output=../../../mocks --filename=mock_broker.go --with-expecter
type Broker interface {
	StartConsumer(ctx context.Context, handler func(context.Context, []byte) error) error
}
