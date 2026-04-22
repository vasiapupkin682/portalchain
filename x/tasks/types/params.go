package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type Params struct {
    FreeTextLimit      int64    `json:"free_text_limit"`
    FreeCodeLimit      int64    `json:"free_code_limit"`
    FreeAnalysisLimit  int64    `json:"free_analysis_limit"`
    PricePerText       sdk.Coin `json:"price_per_text"`
    PricePerCode       sdk.Coin `json:"price_per_code"`
    PricePerAnalysis   sdk.Coin `json:"price_per_analysis"`
    TaskDeadlineBlocks int64    `json:"task_deadline_blocks"`
}

func DefaultParams() Params {
    return Params{
        FreeTextLimit:      100,
        FreeCodeLimit:      50,
        FreeAnalysisLimit:  20,
        PricePerText:       sdk.NewCoin("udaai", sdk.NewInt(1)),
        PricePerCode:       sdk.NewCoin("udaai", sdk.NewInt(5)),
        PricePerAnalysis:   sdk.NewCoin("udaai", sdk.NewInt(10)),
        TaskDeadlineBlocks: 100,
    }
}
