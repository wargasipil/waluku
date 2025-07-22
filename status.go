package waluku

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/pdcgo/shared/pkg/debugtool"
)

type ValuationData struct {
	Value float64 `json:"value"`
	Asset float64 `json:"asset"`
	Cash  float64 `json:"cash"`
}

type Report interface {
	GetValuation(ctx context.Context) (*ValuationData, error)
}

type reportImpl struct {
	c *binance.Client
}

// GetValuation implements Report.
func (r *reportImpl) GetValuation(ctx context.Context) (*ValuationData, error) {
	res, _ := r.c.NewGetUserAsset().Do(ctx)

	debugtool.LogJson(res)
	return nil, nil
}

func NewBinanceReport(client *binance.Client) Report {
	return &reportImpl{
		c: client,
	}
}
