package vite

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/coinbase/rosetta-sdk-go/storage/modules"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/coinbase/rosetta-sdk-go/utils"
	viteTypes "github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/config"
)

type genesis struct {
	GenesisAccountAddress 	string 					   	`json:"GenesisAccountAddress"`
	ForkPoints				[]interface{}				`json:"ForkPoints"`
	GovernanceInfo			map[string]interface{}		`json:"GovernanceInfo"`
	AssetInfo       		assetInfo                  	`json:"AssetInfo"`
	QuotaInfo 				map[string]interface{}		`json:"QuotaInfo"`
	AccountBalanceMap 		map[string]accountBalances 	`json:"AccountBalanceMap"`
}

type assetInfo struct {
	TokenInfoMap 	map[string]config.TokenInfo `json:"TokenInfoMap"`
	LogList			[]interface{} 				`json:"LogList"`
}

type accountBalances map[string]big.Int

func viteTokenToCurrency(tti string, tokenInfo config.TokenInfo) (currency types.Currency) {
	currency = types.Currency{
		Symbol:   tokenInfo.TokenSymbol,
		Decimals: int32(tokenInfo.Decimals),
		Metadata: map[string]interface{}{
			"tti": tti,
		},
	}
	return
}

// GenerateBootstrapFile creates the bootstrap balances file
// for a particular genesis file.
func GenerateBootstrapFile(genesisFile string, outputFile string) error {
	var genesisAllocations genesis
	if err := utils.LoadAndParse(genesisFile, &genesisAllocations); err != nil {
		return fmt.Errorf("%w: could not load genesis file", err)
	}

	balances := []*modules.BootstrapBalance{}
	for address := range genesisAllocations.AccountBalanceMap {
		_, err := viteTypes.HexToAddress(address)
		if err != nil {
			return fmt.Errorf("invalid address 0x%s", address)
		}
	
		accountBalances := genesisAllocations.AccountBalanceMap[address]
		for tti, value := range accountBalances {
			// Skip zero balances
			if len(value.Bits()) == 0 {
				continue
			}
			tokenInfo := genesisAllocations.AssetInfo.TokenInfoMap[tti]
			currency := viteTokenToCurrency(tti, tokenInfo)
			balances = append(balances, &modules.BootstrapBalance{
				Account: &types.AccountIdentifier{
					Address: address,
				},
				Value: value.String(),
				Currency: &currency,
			})
		}
	}
	sort.Slice(balances, func(i, j int) bool {
		first := balances[i]
		second := balances[j]

		if first.Account.Address == second.Account.Address {
			return first.Currency.Symbol < second.Currency.Symbol
		}
		return first.Account.Address < second.Account.Address
	})
	

	// Write to file
	if err := utils.SerializeAndWrite(outputFile, balances); err != nil {
		return fmt.Errorf("%w: could not write bootstrap balances", err)
	}

	return nil
}
