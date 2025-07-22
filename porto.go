package waluku

import (
	"log"
	"log/slog"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/pdcgo/shared/yenstream"
)

type PortoState struct {
	Asset  float64
	Quoted float64
}

func NewPortoCalc(client *binance.Client, ctx *yenstream.RunnerContext, pipe yenstream.Pipeline) (yenstream.Pipeline, func() error, error) {
	st := &PortoState{}

	freshPorto := func() error {
		time.Sleep(time.Second * 2)
		slog.Info("refresh porto")
		var err error
		res, err := client.NewGetUserAsset().Asset("BNB").Do(ctx)
		if err != nil {
			return err
		}

		st.Asset, err = strconv.ParseFloat(res[0].Free, 64)
		if err != nil {
			return err
		}

		res, err = client.NewGetUserAsset().Asset("USDT").Do(ctx)
		if err != nil {
			return err
		}

		st.Quoted, err = strconv.ParseFloat(res[0].Free, 64)
		if err != nil {
			return err

		}

		return nil
	}

	err := freshPorto()
	if err != nil {
		return pipe, freshPorto, err
	}

	return pipe.
		Via("calculate_asset", yenstream.NewMap(ctx, func(data *binance.WsKlineEvent) (*binance.WsKlineEvent, error) {

			currPrice, err := strconv.ParseFloat(data.Kline.Close, 64)
			if err != nil {
				return data, err
			}
			value := st.Quoted + (st.Asset * currPrice)
			log.Printf("Value %.5f asset %.5f cash %.5f --> %.5f\n", value, (st.Asset * currPrice), st.Quoted, currPrice)

			return data, nil
		})), freshPorto, nil
}
