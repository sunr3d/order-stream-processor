package infra

import "context"

//go:generate go run github.com/vektra/mockery/v2@v2.53.2 --name=Broker --output=../../../mocks --filename=mock_broker.go --with-expecter
type Broker interface {
	Start(ctx context.Context) error
	Stop() error
}
