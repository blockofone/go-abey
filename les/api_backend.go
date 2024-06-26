// Copyright 2016 The go-ethereum Authors
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

package les

import (
	"context"
	"errors"
	"github.com/AbeyFoundation/go-abey/abey/downloader"
	"github.com/AbeyFoundation/go-abey/abey/fastdownloader"
	"github.com/AbeyFoundation/go-abey/light"
	"math/big"

	"github.com/AbeyFoundation/go-abey/abey/gasprice"
	"github.com/AbeyFoundation/go-abey/abeydb"
	"github.com/AbeyFoundation/go-abey/accounts"
	"github.com/AbeyFoundation/go-abey/common"
	"github.com/AbeyFoundation/go-abey/common/math"
	"github.com/AbeyFoundation/go-abey/core"
	"github.com/AbeyFoundation/go-abey/core/bloombits"
	"github.com/AbeyFoundation/go-abey/core/rawdb"
	"github.com/AbeyFoundation/go-abey/core/state"
	"github.com/AbeyFoundation/go-abey/core/types"
	"github.com/AbeyFoundation/go-abey/core/vm"
	"github.com/AbeyFoundation/go-abey/event"
	"github.com/AbeyFoundation/go-abey/params"
	"github.com/AbeyFoundation/go-abey/rpc"
)

type LesApiBackend struct {
	abey *LightAbey
	gpo  *gasprice.Oracle
}

var (
	NotSupportOnLes = errors.New("not support on les protocol")
)

// ////////////////////////////////////////////////////////////
func (b *LesApiBackend) SetSnailHead(number uint64) {

}
func (b *LesApiBackend) SnailHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.SnailHeader, error) {
	return nil, NotSupportOnLes
}
func (b *LesApiBackend) SnailBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.SnailBlock, error) {
	return nil, NotSupportOnLes
}
func (b *LesApiBackend) GetFruit(ctx context.Context, fastblockHash common.Hash) (*types.SnailBlock, error) {
	return nil, NotSupportOnLes
}
func (b *LesApiBackend) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *types.Header, error) {
	return nil, nil, NotSupportOnLes
}
func (b *LesApiBackend) StateAndHeaderByHash(ctx context.Context, hash common.Hash) (*state.StateDB, *types.Header, error) {
	return nil, nil, NotSupportOnLes
}
func (b *LesApiBackend) GetSnailBlock(ctx context.Context, blockHash common.Hash) (*types.SnailBlock, error) {
	return nil, NotSupportOnLes
}
func (b *LesApiBackend) GetReward(number int64) *types.BlockReward {
	return nil
}
func (b *LesApiBackend) GetCommittee(id rpc.BlockNumber) (map[string]interface{}, error) {
	return nil, NotSupportOnLes
}
func (b *LesApiBackend) GetCurrentCommitteeNumber() *big.Int {
	return nil
}
func (b *LesApiBackend) GetStateChangeByFastNumber(fastNumber rpc.BlockNumber) *types.BlockBalance {
	return nil
}
func (b *LesApiBackend) GetBalanceChangeBySnailNumber(snailNumber rpc.BlockNumber) *types.BalanceChangeContent {
	return nil
}
func (b *LesApiBackend) GetSnailRewardContent(blockNr rpc.BlockNumber) *types.SnailRewardContenet {
	return nil
}
func (b *LesApiBackend) GetChainRewardContent(blockNr rpc.BlockNumber) *types.ChainReward {
	return nil
}
func (b *LesApiBackend) CurrentSnailBlock() *types.SnailBlock {
	return nil
}
func (b *LesApiBackend) SnailPoolContent() []*types.SnailBlock {
	return nil
}
func (b *LesApiBackend) SnailPoolInspect() []*types.SnailBlock {
	return nil
}
func (b *LesApiBackend) SnailPoolStats() (pending int, unVerified int) {
	return 0, 0
}
func (b *LesApiBackend) Downloader() *downloader.Downloader {
	return nil
}

// ////////////////////////////////////////////////////////////
func (b *LesApiBackend) ChainConfig() *params.ChainConfig {
	return b.abey.chainConfig
}

func (b *LesApiBackend) CurrentBlock() *types.Block {
	return types.NewBlockWithHeader(b.abey.blockchain.CurrentHeader())
}

func (b *LesApiBackend) SetHead(number uint64) {
	b.abey.protocolManager.downloader.Cancel()
	b.abey.blockchain.SetHead(number)
}

func (b *LesApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	if blockNr == rpc.LatestBlockNumber || blockNr == rpc.PendingBlockNumber {
		return b.abey.blockchain.CurrentHeader(), nil
	}

	return b.abey.blockchain.GetHeaderByNumberOdr(ctx, uint64(blockNr))
}
func (b *LesApiBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.abey.blockchain.GetHeaderByHash(hash), nil
}

func (b *LesApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, err
	}
	return b.GetBlock(ctx, header.Hash())
}

func (b *LesApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	return light.NewState(ctx, header, b.abey.odr), header, nil
}

func (b *LesApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.abey.blockchain.GetBlockByHash(ctx, blockHash)
}

func (b *LesApiBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.abey.chainDb, hash); number != nil {
		return light.GetBlockReceipts(ctx, b.abey.odr, hash, *number)
	}
	return nil, nil
}

func (b *LesApiBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	if number := rawdb.ReadHeaderNumber(b.abey.chainDb, hash); number != nil {
		return light.GetBlockLogs(ctx, b.abey.odr, hash, *number)
	}
	return nil, nil
}

func (b *LesApiBackend) GetTd(hash common.Hash) *big.Int {
	return big.NewInt(0)
	//return b.abey.blockchain.GetTdByHash(hash)
}

func (b *LesApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	context := core.NewEVMContext(msg, header, b.abey.blockchain, nil, nil)
	return vm.NewEVM(context, state, b.abey.chainConfig, vmCfg), state.Error, nil
}

func (b *LesApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.abey.txPool.Add(ctx, signedTx)
}

func (b *LesApiBackend) RemoveTx(txHash common.Hash) {
	b.abey.txPool.RemoveTx(txHash)
}

func (b *LesApiBackend) GetPoolTransactions() (types.Transactions, error) {
	return b.abey.txPool.GetTransactions()
}

func (b *LesApiBackend) GetPoolTransaction(txHash common.Hash) *types.Transaction {
	return b.abey.txPool.GetTransaction(txHash)
}

func (b *LesApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.abey.txPool.GetNonce(ctx, addr)
}

func (b *LesApiBackend) Stats() (pending int, queued int) {
	return b.abey.txPool.Stats(), 0
}

func (b *LesApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.abey.txPool.Content()
}

func (b *LesApiBackend) SubscribeNewTxsEvent(ch chan<- types.NewTxsEvent) event.Subscription {
	return b.abey.txPool.SubscribeNewTxsEvent(ch)
}

func (b *LesApiBackend) SubscribeChainEvent(ch chan<- types.FastChainEvent) event.Subscription {
	return b.abey.blockchain.SubscribeChainEvent(ch)
}

func (b *LesApiBackend) SubscribeChainHeadEvent(ch chan<- types.FastChainHeadEvent) event.Subscription {
	return b.abey.blockchain.SubscribeChainHeadEvent(ch)
}

func (b *LesApiBackend) SubscribeChainSideEvent(ch chan<- types.FastChainSideEvent) event.Subscription {
	return b.abey.blockchain.SubscribeChainSideEvent(ch)
}

func (b *LesApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.abey.blockchain.SubscribeLogsEvent(ch)
}

func (b *LesApiBackend) SubscribeRemovedLogsEvent(ch chan<- types.RemovedLogsEvent) event.Subscription {
	return b.abey.blockchain.SubscribeRemovedLogsEvent(ch)
}

func (b *LesApiBackend) FastDownloader() *fastdownloader.Downloader {
	return b.abey.Downloader()
}

func (b *LesApiBackend) ProtocolVersion() int {
	return b.abey.LesVersion() + 10000
}

func (b *LesApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *LesApiBackend) ChainDb() abeydb.Database {
	return b.abey.chainDb
}

func (b *LesApiBackend) EventMux() *event.TypeMux {
	return b.abey.eventMux
}

func (b *LesApiBackend) AccountManager() *accounts.Manager {
	return b.abey.accountManager
}

func (b *LesApiBackend) BloomStatus() (uint64, uint64) {
	if b.abey.bloomIndexer == nil {
		return 0, 0
	}
	sections, _, _ := b.abey.bloomIndexer.Sections()
	return params.BloomBitsBlocksClient, sections
}

func (b *LesApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.abey.bloomRequests)
	}
}
