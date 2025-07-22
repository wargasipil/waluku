package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/pdcgo/shared/pkg/debugtool"
	"github.com/pdcgo/shared/yenstream"
	"github.com/pdcgo/shared/yenstream/store"
	"github.com/wargasipil/waluku"
)

var (
	apiKey    = ""
	secretKey = ""
)
var AppName = "waluku"
var INTERVAL = "1m"
var SYMBOL = waluku.SymbolPair{"BNB", "USDT"}

type State struct {
	CurrentPrice float64 `json:"current_price"`
}

func (s *State) SetCurrentPrice(data string) error {
	c, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return err
	}

	s.CurrentPrice = c
	return nil
}

func main() {
	pctx := context.Background()
	client := binance.NewClient(apiKey, secretKey)
	trader := waluku.NewBinanceTrader(client, SYMBOL)
	ctx, _, err := store.RegisterBadgerStore(pctx, AppName)
	if err != nil {
		log.Panicln(err)
	}

	kline := make(chan any, 1)
	go func() {
		defer close(kline)
		// errChan := make(chan error, 1)

	Loop:
		for {
			// feeding kline
			doneC, stopC, err := binance.WsKlineServe(
				SYMBOL.String(),
				INTERVAL,
				func(event *binance.WsKlineEvent) {
					kline <- event
				},
				func(err error) {
					slog.Error(err.Error(), slog.String("source", "feed"))
				},
			)

			if err != nil {
				slog.Error(err.Error())
				time.Sleep(time.Second * 5)
				log.Println("retry")
				continue
			}

			select {
			case <-doneC:
				continue
			case <-stopC:
				continue
			case <-ctx.Done():
				break Loop

			}
			// <-stop
		}

	}()

	current_price := float64(0)

	yenstream.
		NewRunnerContext(ctx).
		CreatePipeline(func(ctx *yenstream.RunnerContext) yenstream.Pipeline {
			source := yenstream.NewChannelSource(ctx, kline)
			kline := source.
				Via("update_to_current_price", yenstream.NewMap(ctx, func(data *binance.WsKlineEvent) (*binance.WsKlineEvent, error) {
					currPrice, err := strconv.ParseFloat(data.Kline.Close, 64)
					if err != nil {
						return data, err
					}

					current_price = currPrice
					return data, nil
				})).
				Via("get_true_kline", yenstream.NewFilter(ctx, func(data *binance.WsKlineEvent) (bool, error) {
					return data.Kline.IsFinal, nil
				}))

			indicator := SwingIndicator(ctx, 48, kline).
				Via("debug_indicator", yenstream.NewMap(ctx, debugMap)).
				Via("filter_new_hl", yenstream.NewFilter(ctx, func(data *SwingIndicatorData) (bool, error) {
					return data.IsHigh || data.IsLow, nil
				}))

			porto, freshPorto, err := waluku.NewPortoCalc(
				client,
				ctx,
				source.Via("empty_map_porto", yenstream.NewMap(ctx, emptyMap)),
			)
			if err != nil {
				panic(err)
			}

			tradeops := indicator.
				Via("trade", yenstream.NewMap(ctx, func(data *SwingIndicatorData) (*SwingIndicatorData, error) {
					var err error
					var pos *waluku.Position
					pos, err = trader.Position(ctx)
					if err != nil {
						return nil, err
					}

					if pos != nil {
						now := (pos.Qty * current_price)
						last := (pos.Qty * pos.Price)
						margin := now - last
						debugtool.LogJson(pos)
						log.Printf("Margin %.5f, now %.5f, past %.5f\n", margin, now, last)

						if margin <= 0.005 {
							return data, nil
						}

						if data.IsHigh {
							err := trader.Sell(ctx, fmt.Sprintf("%.8f", current_price))
							if err != nil {
								return data, err
							}
							err = freshPorto()
							if err != nil {
								return data, err
							}
							return data, err
						}

						return data, nil
					}

					if !data.IsLow {
						err = trader.Buy(ctx, fmt.Sprintf("%.8f", current_price))
						err := freshPorto()
						if err != nil {
							return data, err
						}
						return data, err
					}

					return data, nil
				}))

			// current := source.
			// 	Via("parsing_state", yenstream.NewMap(ctx, func(data *binance.WsKlineEvent) (*State, error) {
			// 		current_price := data.Kline.Close
			// 		err := state.SetCurrentPrice(current_price)
			// 		if err != nil {
			// 			return nil, err
			// 		}

			// 		return &state, nil
			// 	})).
			// 	Via("debug_live", yenstream.NewMap(ctx, debugMap))

			return yenstream.NewFlatten(ctx, "flatten_end", tradeops, porto)
		})

}

func emptyMap(data any) (any, error) {
	return data, nil
}

func debugMap(data any) (any, error) {
	raw, err := json.Marshal(data)
	log.Println(string(raw))
	return data, err
}
