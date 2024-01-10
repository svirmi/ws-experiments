package utils

import (
	"context"

	bnspotmd "github.com/linstohu/nexapi/binance/spot/marketdata"
	bnspottypes "github.com/linstohu/nexapi/binance/spot/marketdata/types"
	bnspotutils "github.com/linstohu/nexapi/binance/spot/utils"
)

// const TRADING = "BREAK"
const TRADING = "TRADING"

type tPair struct {
	Symbol     string
	QuoteAsset string
}

func TradablePairs(quoteAsset string) []tPair {

	var tPairs []tPair

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	cli, err := bnspotmd.NewSpotMarketDataClient(&bnspotutils.SpotClientCfg{
		Debug:   false,
		BaseURL: bnspotutils.BaseURL,
	})
	if err != nil {
		panic(err)
	}

	exhInfo, err := cli.GetExchangeInfo(ctx, bnspottypes.GetExchangeInfoParam{})

	if err != nil {
		panic(err)
	}

	for _, pair := range exhInfo.Body.Symbols {
		if pair.Status == TRADING && pair.IsSpotTradingAllowed && pair.QuoteAsset == quoteAsset {
			tPairs = append(tPairs, tPair{Symbol: pair.Symbol, QuoteAsset: pair.QuoteAsset})
		}
	}

	return tPairs
}

// TODO : calculate market cap
// https://api.binance.com/api/v3/ticker/24hr?symbols=[%22ETHUSDT%22%2C%22LTCBTC%22]
// https://dev.binance.vision/t/get-market-cap-api/3050

// check combined streams
// https://dev.binance.vision/t/question-single-raw-stream-or-a-combined-stream/2997
