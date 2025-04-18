// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	//"github.com/AbeyFoundation/go-abey/common"
	"github.com/AbeyFoundation/go-abey/crypto"
	"github.com/AbeyFoundation/go-abey/metrics"
	"math"
	"time"

	//"github.com/AbeyFoundation/go-abey/log"
	"github.com/AbeyFoundation/go-abey/consensus"
	"github.com/AbeyFoundation/go-abey/core/state"
	"github.com/AbeyFoundation/go-abey/core/types"
	"github.com/AbeyFoundation/go-abey/core/vm"
	"github.com/AbeyFoundation/go-abey/params"

	"math/big"
)

var (
	blockExecutionTxTimer = metrics.NewRegisteredTimer("chain/state/executiontx", nil)
	blockFinalizeTimer    = metrics.NewRegisteredTimer("chain/state/finalize", nil)
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (fp *StateProcessor) Process(block *types.Block, statedb *state.StateDB,
	cfg vm.Config) (types.Receipts, []*types.Log, uint64, *types.ChainReward, error) {
	var (
		receipts  types.Receipts
		usedGas   = new(uint64)
		feeAmount = big.NewInt(0)
		header    = block.Header()
		allLogs   []*types.Log
		gp        = new(GasPool).AddGas(block.GasLimit())
	)
	start := time.Now()
	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		txhash := tx.HashOld()
		if fp.config.IsTIP10(block.Number()) {
			txhash = tx.Hash()
		}
		statedb.Prepare(txhash, block.Hash(), i)
		receipt, err := ApplyTransaction(fp.config, fp.bc, gp, statedb, header, tx, usedGas, feeAmount, cfg)
		if err != nil {
			return nil, nil, 0, nil, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}
	t1 := time.Now()
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	_, infos, err := fp.engine.Finalize(fp.bc, header, statedb, block.Transactions(), receipts, feeAmount)
	if err != nil {
		return nil, nil, 0, nil, err
	}
	blockExecutionTxTimer.Update(t1.Sub(start))
	blockFinalizeTimer.Update(time.Since(t1))
	return receipts, allLogs, *usedGas, infos, nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc ChainContext, gp *GasPool,
	statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, feeAmount *big.Int, cfg vm.Config) (*types.Receipt, error) {
	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return nil, err
	}
	if header.Number.Cmp(big.NewInt(6638000)) > 0 {
		if err := types.ForbidAddress(msg.From()); err != nil {
			return nil, err
		}

		if header.Number.Cmp(big.NewInt(24642000)) > 0 {
			if err := types.ForbidAddress2(msg.From()); err != nil {
				return nil, err
			}
		}
	}

	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msg, header, bc, nil, nil)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, statedb, config, cfg)
	// Apply the transaction to the current state (included in the env)
	result, err := ApplyMessage(vmenv, msg, gp)

	if err != nil {
		return nil, err
	}
	// Update the state with pending changes
	var root []byte

	statedb.Finalise(true)

	*usedGas += result.UsedGas
	gasFee := new(big.Int).Mul(new(big.Int).SetUint64(result.UsedGas), msg.GasPrice())
	feeAmount.Add(gasFee, feeAmount)
	if msg.Fee() != nil {
		feeAmount.Add(msg.Fee(), feeAmount) //add fee
	}
	txhash := tx.HashOld()
	if config.IsTIP10(header.Number) {
		txhash = tx.Hash()
	}
	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	receipt := types.NewReceipt(root, result.Failed(), *usedGas)
	receipt.TxHash = txhash
	receipt.GasUsed = result.UsedGas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, tx.Nonce())
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(txhash)
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
	receipt.BlockHash = statedb.BlockHash()
	receipt.BlockNumber = header.Number
	receipt.TransactionIndex = uint(statedb.TxIndex())

	return receipt, err
}

// ReadTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the result
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ReadTransaction(config *params.ChainConfig, bc ChainContext,
	statedb *state.StateDB, header *types.Header, tx *types.Transaction, cfg vm.Config) ([]byte, uint64, error) {

	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))

	msgCopy := types.NewMessage(msg.From(), msg.To(), msg.Payment(), 0, msg.Value(), msg.Fee(), msg.Gas(), msg.GasPrice(), msg.Data(), false)

	if err != nil {
		return nil, 0, err
	}
	if header.Number.Cmp(big.NewInt(6638000)) > 0 {
		if err := types.ForbidAddress(msgCopy.From()); err != nil {
			return nil, 0, err
		}
	}

	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msgCopy, header, bc, nil, nil)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, statedb, config, cfg)
	// Apply the transaction to the current state (included in the env)
	gp := new(GasPool).AddGas(math.MaxUint64)
	result, err := ApplyMessage(vmenv, msg, gp)
	if err != nil {
		return nil, 0, err
	}

	return result.ReturnData, result.UsedGas, err
}
