package services

import (
	"net/http"

	"github.com/azbuky/rosetta-vite/configuration"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/server"
)

// NewBlockchainRouter creates a Mux http.Handler from a collection
// of server controllers.
func NewBlockchainRouter(
	config *configuration.Configuration,
	client Client,
	asserter *asserter.Asserter,
) http.Handler {
	networkAPIService := NewNetworkAPIService(config, client)
	networkAPIController := server.NewNetworkAPIController(
		networkAPIService,
		asserter,
	)

	accountAPIService := NewAccountAPIService(config, client)
	accountAPIController := server.NewAccountAPIController(
		accountAPIService,
		asserter,
	)

	blockAPIService := NewBlockAPIService(config, client)
	blockAPIController := server.NewBlockAPIController(
		blockAPIService,
		asserter,
	)

	constructionAPIService := NewConstructionAPIService(config, client)
	constructionAPIController := server.NewConstructionAPIController(
		constructionAPIService,
		asserter,
	)

	return server.NewRouter(
		networkAPIController,
		accountAPIController,
		blockAPIController,
		constructionAPIController,
	)
}
