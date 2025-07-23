package mocks

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type MockBadAPI struct{}

func (m *MockBadAPI) Chain() (types.Text, error) {
	return types.Text("WrongChain"), nil
}

func (m *MockBadAPI) Name() (types.Text, error) {
	return types.Text("UnexpectedNode"), nil
}

func (m *MockBadAPI) Version() (types.Text, error) {
	return types.Text("v0.0.1"), nil
}

func (m *MockBadAPI) GetBlockHashLatest() (types.Hash, error) {
	var h types.Hash
	copy(h[:], []byte("wronghash000000000000000000000000"))
	return h, nil
}
