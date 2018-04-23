package ethfw

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/AtlantPlatform/ethfw/sol"
)

var (
	ErrNoContract        = errors.New("contract not provided")
	ErrContractNoAddress = errors.New("contract has no address")
	ErrNoPrivateKey      = errors.New("private key not found")
	ErrAlreadyDeployed   = errors.New("not neccesary binary: contract deployed already")
)

type Client interface {
	bind.ContractCaller
	bind.ContractTransactor
	bind.ContractFilterer

	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	TransactOpts(ctx context.Context, account common.Address, password string) (*bind.TransactOpts, error)
	BindContract(contract *sol.Contract) (*BoundContract, error)
	PersonalAccounts(ctx context.Context) ([]common.Address, error)
	Transaction(ctx context.Context, id common.Hash) (*types.Transaction, bool, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	Close()
}

func NewClient(cli *rpc.Client, opts ...clientOpt) Client {
	c := &client{
		cli:        cli,
		eth:        ethclient.NewClient(cli),
		keyCache:   NewKeyCache(),
		nonceCache: NewNonceCache(),
		opt: &clientOptions{
			accountSyncTimeout: 30 * time.Second,
			chainID:            1,
		},
		blockC:   make(chan Block, 1024),
		closeC:   make(chan struct{}),
		blockMux: new(sync.RWMutex),
	}
	for _, o := range opts {
		if o != nil {
			o(c.opt)
		}
	}
	return c
}

type Block struct {
	Number *hexutil.Big
}

type client struct {
	opt *clientOptions

	cli        *rpc.Client
	eth        *ethclient.Client
	keyCache   KeyCache
	nonceCache NonceCache
	blockC     chan Block
	block      *big.Int
	blockMux   *sync.RWMutex
	closeC     chan struct{}
	doneC      chan struct{}
}

func (c *client) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()

	return c.eth.CodeAt(ctx, contract, blockNumber)
}

func (c *client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()

	return c.eth.BalanceAt(ctx, account, blockNumber)
}

func (c *client) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()

	return c.eth.CallContract(ctx, call, blockNumber)
}

func (c *client) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()

	return c.eth.PendingCodeAt(ctx, account)
}

func (c *client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()

	return c.eth.PendingNonceAt(ctx, account)
}

func (c *client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()

	return c.eth.SuggestGasPrice(ctx)
}

func (c *client) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()

	return c.eth.EstimateGas(ctx, call)
}

func (c *client) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()
	return c.SendTransaction(ctx, tx)
}

func (c *client) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()

	return c.eth.FilterLogs(ctx, query)
}

func (c *client) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()

	return c.eth.SubscribeFilterLogs(ctx, query, ch)
}

func (c *client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()

	return c.eth.TransactionReceipt(ctx, txHash)
}

func (c *client) TransactOpts(ctx context.Context, account common.Address, password string) (*bind.TransactOpts, error) {
	signFunc := c.keyCache.SignerFn(account, password)
	if signFunc == nil {
		return nil, ErrNoPrivateKey
	}
	opts := bind.TransactOpts{
		From:    account,
		Signer:  signFunc,
		Context: ctx,
	}
	return &opts, nil
}

func (c *client) BindContract(contract *sol.Contract) (*BoundContract, error) {
	if contract == nil {
		return nil, ErrNoContract
	} else if len(contract.Address) == 0 && contract.Bin == "" {
		return nil, ErrContractNoAddress
	} else if len(contract.Address) != 0 && len(contract.Bin) != 0 {
		return nil, ErrAlreadyDeployed
	}
	parsedABI, err := abi.JSON(bytes.NewReader(contract.ABI))
	if err != nil {
		err = fmt.Errorf("failed to parse contract ABI: %v", err)
		return nil, err
	}
	bound := &BoundContract{
		BoundContract: bind.NewBoundContract(contract.Address, parsedABI, c, c, c),
		address:       contract.Address,
		cli:           c,
		abi:           parsedABI,
		src:           contract,
	}
	return bound, nil
}

func (c *client) PersonalAccounts(ctx context.Context) (result []common.Address, err error) {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()
	err = c.cli.CallContext(ctx, &result, "eth_accounts", nil)
	if err != nil {
		err = fmt.Errorf("accounts weren't found: %v", err)
		return nil, err
	}
	return result, nil
}

func (c *client) Transaction(ctx context.Context, id common.Hash) (*types.Transaction, bool, error) {
	ctx, cancelFn := ContextWithCloseChan(ctx, c.closeC)
	defer cancelFn()
	tx, isPending, err := c.eth.TransactionByHash(ctx, id)
	if err != nil {
		return nil, isPending, fmt.Errorf("failed to get tx by hash: %v", err)
	}
	return tx, isPending, nil

}

func (c *client) Close() {
	close(c.closeC)
	<-c.doneC
}
