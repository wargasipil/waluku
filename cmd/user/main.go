package main

import (
	"context"
	"log"

	"github.com/adshao/go-binance/v2"
	"github.com/pdcgo/shared/pkg/debugtool"
)

var (
	apiKey    = ""
	secretKey = ""
)

func main() {
	ctx := context.Background()
	client := binance.NewClient(apiKey, secretKey)

	// res, err := client.NewCreateOrderService().
	// 	Side(binance.SideTypeBuy).
	// 	Type(binance.OrderTypeLimit).
	// 	Symbol("BNBUSDT").
	// 	TimeInForce(binance.TimeInForceTypeGTC).
	// 	Price("780.00").
	// 	Quantity("").
	// 	// QuoteOrderQty("5").
	// 	Do(ctx)

	// if err != nil {
	// 	panic(err)

	// }

	// debugtool.LogJson(res)

	// res, err := client.NewKlinesService().
	// 	Symbol("BNBUSDT").
	// 	Interval("1m").
	// 	StartTime(time.Now().Add(time.Hour * 24).Unix()).
	// 	Do(context.Background())

	// if err != nil {
	// 	panic(err)
	// }

	// for _, kline := range res {
	// 	log.Println(kline.CloseTime)
	// }

	// res[0].

	res, err := client.NewExchangeInfoService().Symbol("BNBUSDT").Do(ctx)
	log.Println(res, err)
	debugtool.LogJson(res)

	// res, err := client.NewGetAllCoinsInfoService().Do(context.Background())
	// log.Println(res, err)

	// for _, coin := range res {
	// 	if coin.Coin == "BNB" {
	// 		debugtool.LogJson(coin)
	// 	}
	// }

	// res, err := client.NewCreateOrderService().
	// 	Side(binance.SideTypeBuy).
	// 	Type(binance.OrderTypeMarket).
	// 	Symbol("BNBUSDT").
	// 	QuoteOrderQty("5").
	// 	Do(context.Background())

	// if err != nil {
	// 	panic(err)
	// }
	// debugtool.LogJson(res)

}
