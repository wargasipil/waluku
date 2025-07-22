package main

import (
	"container/ring"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/pdcgo/shared/yenstream"
)

type SwingIndicatorData struct {
	IsLow  bool    `json:"is_low"`
	Low    float64 `json:"low"`
	IsHigh bool    `json:"is_high"`
	High   float64 `json:"high"`
}

type SwingState struct {
	p *ring.Ring
}

func (s *SwingState) Add(d any) {
	p := s.p
	p.Value = d

	s.p = p.Next()

}

func SwingIndicator(ctx *yenstream.RunnerContext, period int, pipe yenstream.Pipeline) yenstream.Pipeline {
	state := SwingState{
		p: ring.New(period),
	}

	client := binance.NewClient(apiKey, secretKey)
	res, err := client.NewKlinesService().
		Symbol(SYMBOL.String()).
		Interval(INTERVAL).
		StartTime(time.Now().Add(time.Minute * time.Duration(period)).UnixMicro()).
		Do(context.Background())

	if err != nil {
		panic(err)
	}

	for _, dd := range res {
		kline := dd
		// debugtool.LogJson(kline)
		state.Add(kline)
	}

	return pipe.
		Via("mapping_indicator", yenstream.NewMap(ctx, func(data *binance.WsKlineEvent) (*SwingIndicatorData, error) {
			curHigh, curLow, err := getHighLow(data)
			if err != nil {
				return nil, err
			}

			var high float64 = 0
			var low float64 = 0

			state.p.Do(func(a any) {
				if a == nil {
					return
				}

				var err error
				var chigh, clow float64
				chigh, clow, err = getHighLow(a)

				if chigh > high {
					high = chigh
				}

				if clow < low || low == 0 {
					low = clow
				}
				if err != nil {
					panic(err)
				}

			})

			indData := SwingIndicatorData{
				IsLow:  curLow < low,
				Low:    low,
				IsHigh: curHigh > high,
				High:   high,
			}

			if indData.IsLow {
				indData.Low = curLow
			}

			if indData.IsHigh {
				indData.High = curHigh
			}

			state.Add(data)

			return &indData, nil
		}))
}

func getHighLow(data any) (float64, float64, error) {

	switch kline := data.(type) {
	case *binance.Kline:
		curHigh, _ := strconv.ParseFloat(kline.High, 64)
		curLow, _ := strconv.ParseFloat(kline.Low, 64)
		return curHigh, curLow, nil
	case *binance.WsKlineEvent:
		curHigh, _ := strconv.ParseFloat(kline.Kline.High, 64)
		curLow, _ := strconv.ParseFloat(kline.Kline.Low, 64)
		return curHigh, curLow, nil
	default:
		return 0, 0, errors.New("not type kline or ws klne")
	}
}
