package substrate

import (
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// substrate/api_wrapper.go
type APIWrapper struct {
	API *gsrpc.SubstrateAPI
}

func (w *APIWrapper) Chain() (types.Text, error) {
	return w.API.RPC.System.Chain()
}

func (w *APIWrapper) Name() (types.Text, error) {
	return w.API.RPC.System.Name()
}

func (w *APIWrapper) Version() (types.Text, error) {
	return w.API.RPC.System.Version()
}

func (w *APIWrapper) GetBlockHashLatest() (types.Hash, error) {
	return w.API.RPC.Chain.GetBlockHashLatest()
}
