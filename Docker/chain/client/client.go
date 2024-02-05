package rpc

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
)

func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	r.HandleFunc("/bandchain/chain_id", GetChainIDFn(cliCtx)).Methods("GET")
	r.HandleFunc("/bandchain/genesis", GetGenesisHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/bandchain/evm-validators", GetEVMValidators(cliCtx)).Methods("GET")

	r.HandleFunc("/bandchain/v1/chain_id", GetChainIDFn(cliCtx)).Methods("GET")
	r.HandleFunc("/bandchain/v1/evm-validators", GetEVMValidators(cliCtx)).Methods("GET")
}
