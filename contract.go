// Copyright 2017, 2018 Tensigma Ltd. All rights reserved.
// Use of this source code is governed by Microsoft Reference Source
// License (MS-RSL) that can be found in the LICENSE file.

package ethfw

import (
	"fmt"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/AtlantPlatform/ethfw/sol"
)

type BoundContract struct {
	*bind.BoundContract
	address common.Address
	src     *sol.Contract
	abi     abi.ABI
	cli     *client
}

func (c *BoundContract) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	input, err := c.abi.Pack(method, params...)
	if err != nil {
		return nil, err
	}
	return c.transact(opts, &c.address, input)
}

// Invokes PAID contract method
func (c *BoundContract) transact(
	opts *bind.TransactOpts,
	contract *common.Address,
	input []byte,
) (
	*types.Transaction,
	error,
) {
	var err error
	ctx := opts.Context
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	var nonce uint64
	nonceCache := c.cli.nonceCache
	if opts.Nonce == nil {
		nonce = nonceCache.Get(opts.From)
		if nonce == 0 {
			nonce = 1
			nonceCache.Set(opts.From, nonce)
		} else {
			nonce = nonceCache.Incr(opts.From)
		}
	} else {
		nonce = opts.Nonce.Uint64()
	}
	gasPrice := opts.GasPrice
	if gasPrice == nil {
		gasPrice, err = c.cli.SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to suggest gas price: %v", err)
		}
	}
	gasLimit := opts.GasLimit
	if gasLimit == 0 {
		// Gas estimation cannot succeed without deployed contract
		// Maybe it should be removed
		if contract != nil {
			if code, err := c.cli.PendingCodeAt(ctx, c.address); err != nil {
				return nil, err
			} else if len(code) == 0 {
				return nil, fmt.Errorf("no contract code at given address: %s", c.address.String())
			}
		}
		// If the contract surely has code (or code is not needed), estimate the transaction
		msg := ethereum.CallMsg{From: opts.From, To: contract, Value: value, Data: input}
		gasLimit, err = c.cli.EstimateGas(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to get estimate gas: %v", err)
		}
	}
	// Create the transaction, sign it and schedule it for execution
	var rawTx *types.Transaction
	if contract == nil {
		rawTx = types.NewContractCreation(nonce, value, gasLimit, gasPrice, input)
	} else {
		rawTx = types.NewTransaction(nonce, c.address, value, gasLimit, gasPrice, input)
	}

	if opts.Signer == nil {
		return nil, fmt.Errorf("no signer to authorize the transaction with: %p", opts.Signer)
	}
	signedTx, err := opts.Signer(types.HomesteadSigner{}, opts.From, rawTx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %v", err)
	}
	if err := c.cli.SendTransaction(ctx, signedTx); err != nil {
		nonceCache.Decr(opts.From)
		return nil, err
	}
	return signedTx, nil
}

// NOTE! Executing Transfer on a contract without passing in message data
// is possible for contracts that have a corresponding fallback function. If the
// contract does not have very fallback function, the call will fail
func (c *BoundContract) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return c.transact(opts, &c.address, nil)
}

// Needs testing
func (c *BoundContract) Deploy(opts *bind.TransactOpts, params ...interface{}) (common.Address, *types.Transaction, error) {
	input, err := c.abi.Pack("", params...)
	if err != nil {
		return common.Address{}, nil, err
	}
	tx, err := c.transact(opts, nil, append([]byte(c.src.Bin), input...))
	if err != nil {
		return common.Address{}, nil, err
	}
	c.address = crypto.CreateAddress(opts.From, tx.Nonce())
	return c.address, tx, nil

}
