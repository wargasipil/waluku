package waluku

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/adshao/go-binance/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/pdcgo/shared/pkg/debugtool"
)

type Position struct {
	Qty   float64 `json:"qty"`
	Price float64 `json:"price"`
}

type Trader interface {
	GetValuation(curPrice float64) error
	Buy(ctx context.Context, price string) error
	Sell(ctx context.Context, price string) error
	Position(ctx context.Context) (*Position, error)
	SetPosition(ctx context.Context, pos *Position) error
}

type binanceTraderImpl struct {
	symbol SymbolPair
	c      *binance.Client
	store  TradeStore
}

// Position implements Trader.
func (b *binanceTraderImpl) Position(ctx context.Context) (*Position, error) {
	pos := Position{}
	err := b.store.Get("trade_position", &pos)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, nil
		}
	}
	return &pos, err
}

// SetPosition implements Trader.
func (b *binanceTraderImpl) SetPosition(ctx context.Context, pos *Position) error {
	return b.store.Set("trade_position", pos)
}

// Buy implements Trader.
func (b *binanceTraderImpl) Buy(ctx context.Context, price string) error {
	p, _ := strconv.ParseFloat(price, 64)
	qty := 7.00 / p
	log.Println("buy executed")
	res, err := b.c.NewCreateOrderService().
		Side(binance.SideTypeBuy).
		Type(binance.OrderTypeLimit).
		Symbol(b.symbol.String()).
		TimeInForce(binance.TimeInForceTypeGTC).
		Price(price).
		Quantity(fmt.Sprintf("%.3f", qty)).
		// QuoteOrderQty("5").
		Do(ctx)

	log.Println(fmt.Sprintf("%.3f", qty), "asdasd", price)

	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return errors.New("context cancel on check order")
		default:
			ord, err := b.c.NewGetOrderService().Symbol(b.symbol.String()).OrderID(res.OrderID).Do(ctx)
			if err != nil {
				return err
			}
			debugtool.LogJson(ord)
			if ord.Status == binance.OrderStatusTypeFilled {
				qty, err := strconv.ParseFloat(ord.ExecutedQuantity, 64)
				if err != nil {
					panic(err)
				}

				price, err := strconv.ParseFloat(ord.Price, 64)
				if err != nil {
					panic(err)
				}
				log.Println(ord.Price, ord.ExecutedQuantity)
				err = b.SetPosition(ctx, &Position{
					Qty:   qty,
					Price: price,
				})
				return err
			}
		}
	}
}

// GetValuation implements Trader.
func (b *binanceTraderImpl) GetValuation(curPrice float64) error {
	panic("unimplemented")
}

// Sell implements Trader.
func (b *binanceTraderImpl) Sell(ctx context.Context, price string) error {
	log.Println("sell")
	p, _ := strconv.ParseFloat(price, 64)
	qty := 7.00 / p

	_, err := b.c.NewCreateOrderService().
		Side(binance.SideTypeSell).
		Type(binance.OrderTypeLimit).
		Symbol(b.symbol.String()).
		// QuoteOrderQty("5").
		TimeInForce(binance.TimeInForceTypeGTC).
		Price(price).
		Quantity(fmt.Sprintf("%.3f", qty)).
		Do(ctx)

	if err != nil {
		return err
	}

	return b.store.Delete("trade_position")
}

func NewBinanceTrader(client *binance.Client, symbol SymbolPair) Trader {
	store := NewTradeStore("./.trade_state/" + symbol.String())
	return &binanceTraderImpl{
		symbol: symbol,
		c:      client,
		store:  store,
	}
}
