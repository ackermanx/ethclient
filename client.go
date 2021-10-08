// Package client provides a client for the Ethereum RPC API.
package ethclient

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	erc20 "github.com/ackermanx/ethclient/abi"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
)

// Client defines typed wrappers for the Ethereum RPC API.
type Client struct {
	c          *rpc.Client
	timeout    int
	chainID    *big.Int
	parsedAbis erc20.AddrAbiMap
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*Client, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c *rpc.Client) *Client {
	return &Client{c: c, timeout: 10, parsedAbis: erc20.AddrAbiMap{}}
}

// NewClientWithTimeout creates a client that uses the given RPC client and timeout.
func NewClientWithTimeout(c *rpc.Client, timeout int) *Client {
	return &Client{c: c, timeout: timeout, parsedAbis: erc20.AddrAbiMap{}}
}

func (ec *Client) Close() {
	ec.c.Close()
}

// Blockchain Access

// ChainId retrieves the current chain ID for transaction replay protection.
func (ec *Client) ChainID(ctx context.Context) (*big.Int, error) {
	var result hexutil.Big
	err := ec.c.CallContext(ctx, &result, "eth_chainId")
	if err != nil {
		return nil, err
	}
	return (*big.Int)(&result), err
}

// BlockByHash returns the given full block.
//
// Note that loading full blocks requires two requests. Use HeaderByHash
// if you don't need all transactions or uncle headers.
func (ec *Client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return ec.getBlock(ctx, "eth_getBlockByHash", hash, true)
}

// BlockByNumber returns a block from the current canonical chain. If number is nil, the
// latest known block is returned.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
func (ec *Client) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return ec.getBlock(ctx, "eth_getBlockByNumber", toBlockNumArg(number), true)
}

// BlockNumber returns the most recent block number
func (ec *Client) BlockNumber(ctx context.Context) (uint64, error) {
	var result hexutil.Uint64
	err := ec.c.CallContext(ctx, &result, "eth_blockNumber")
	return uint64(result), err
}

type rpcBlock struct {
	Hash         common.Hash      `json:"hash"`
	Transactions []rpcTransaction `json:"transactions"`
	UncleHashes  []common.Hash    `json:"uncles"`
}

func (ec *Client) getBlock(ctx context.Context, method string, args ...interface{}) (*types.Block, error) {
	var raw json.RawMessage
	err := ec.c.CallContext(ctx, &raw, method, args...)
	if err != nil {
		return nil, err
	} else if len(raw) == 0 {
		return nil, ethereum.NotFound
	}
	// Decode header and transactions.
	var head *types.Header
	var body rpcBlock
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	// Quick-verify transaction and uncle lists. This mostly helps with debugging the server.
	if head.UncleHash == types.EmptyUncleHash && len(body.UncleHashes) > 0 {
		return nil, fmt.Errorf("server returned non-empty uncle list but block header indicates no uncles")
	}
	if head.UncleHash != types.EmptyUncleHash && len(body.UncleHashes) == 0 {
		return nil, fmt.Errorf("server returned empty uncle list but block header indicates uncles")
	}
	if head.TxHash == types.EmptyRootHash && len(body.Transactions) > 0 {
		return nil, fmt.Errorf("server returned non-empty transaction list but block header indicates no transactions")
	}
	if head.TxHash != types.EmptyRootHash && len(body.Transactions) == 0 {
		return nil, fmt.Errorf("server returned empty transaction list but block header indicates transactions")
	}
	// Load uncles because they are not included in the block response.
	var uncles []*types.Header
	if len(body.UncleHashes) > 0 {
		uncles = make([]*types.Header, len(body.UncleHashes))
		reqs := make([]rpc.BatchElem, len(body.UncleHashes))
		for i := range reqs {
			reqs[i] = rpc.BatchElem{
				Method: "eth_getUncleByBlockHashAndIndex",
				Args:   []interface{}{body.Hash, hexutil.EncodeUint64(uint64(i))},
				Result: &uncles[i],
			}
		}
		if err := ec.c.BatchCallContext(ctx, reqs); err != nil {
			return nil, err
		}
		for i := range reqs {
			if reqs[i].Error != nil {
				return nil, reqs[i].Error
			}
			if uncles[i] == nil {
				return nil, fmt.Errorf("got null header for uncle %d of block %x", i, body.Hash[:])
			}
		}
	}
	// Fill the sender cache of transactions in the block.
	txs := make([]*types.Transaction, len(body.Transactions))
	for i, tx := range body.Transactions {
		if tx.From != nil {
			setSenderFromServer(tx.tx, *tx.From, body.Hash)
		}
		txs[i] = tx.tx
	}
	return types.NewBlockWithHeader(head).WithBody(txs, uncles), nil
}

// HeaderByHash returns the block header with the given hash.
func (ec *Client) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	var head *types.Header
	err := ec.c.CallContext(ctx, &head, "eth_getBlockByHash", hash, false)
	if err == nil && head == nil {
		err = ethereum.NotFound
	}
	return head, err
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (ec *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	var head *types.Header
	err := ec.c.CallContext(ctx, &head, "eth_getBlockByNumber", toBlockNumArg(number), false)
	if err == nil && head == nil {
		err = ethereum.NotFound
	}
	return head, err
}

type rpcTransaction struct {
	tx *types.Transaction
	txExtraInfo
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

func (tx *rpcTransaction) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &tx.tx); err != nil {
		return err
	}
	return json.Unmarshal(msg, &tx.txExtraInfo)
}

// TransactionByHash returns the transaction with the given hash.
func (ec *Client) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	var json *rpcTransaction
	err = ec.c.CallContext(ctx, &json, "eth_getTransactionByHash", hash)
	if err != nil {
		return nil, false, err
	} else if json == nil {
		return nil, false, ethereum.NotFound
	} else if _, r, _ := json.tx.RawSignatureValues(); r == nil {
		return nil, false, fmt.Errorf("server returned transaction without signature")
	}
	if json.From != nil && json.BlockHash != nil {
		setSenderFromServer(json.tx, *json.From, *json.BlockHash)
	}
	return json.tx, json.BlockNumber == nil, nil
}

// TransactionSender returns the sender address of the given transaction. The transaction
// must be known to the remote node and included in the blockchain at the given block and
// index. The sender is the one derived by the protocol at the time of inclusion.
//
// There is a fast-path for transactions retrieved by TransactionByHash and
// TransactionInBlock. Getting their sender address can be done without an RPC interaction.
func (ec *Client) TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error) {
	// Try to load the address from the cache.
	sender, err := types.Sender(&senderFromServer{blockhash: block}, tx)
	if err == nil {
		return sender, nil
	}
	var meta struct {
		Hash common.Hash
		From common.Address
	}
	if err = ec.c.CallContext(ctx, &meta, "eth_getTransactionByBlockHashAndIndex", block, hexutil.Uint64(index)); err != nil {
		return common.Address{}, err
	}
	if meta.Hash == (common.Hash{}) || meta.Hash != tx.Hash() {
		return common.Address{}, errors.New("wrong inclusion block/index")
	}
	return meta.From, nil
}

// TransactionCount returns the total number of transactions in the given block.
func (ec *Client) TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error) {
	var num hexutil.Uint
	err := ec.c.CallContext(ctx, &num, "eth_getBlockTransactionCountByHash", blockHash)
	return uint(num), err
}

// TransactionInBlock returns a single transaction at index in the given block.
func (ec *Client) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error) {
	var json *rpcTransaction
	err := ec.c.CallContext(ctx, &json, "eth_getTransactionByBlockHashAndIndex", blockHash, hexutil.Uint64(index))
	if err != nil {
		return nil, err
	}
	if json == nil {
		return nil, ethereum.NotFound
	} else if _, r, _ := json.tx.RawSignatureValues(); r == nil {
		return nil, fmt.Errorf("server returned transaction without signature")
	}
	if json.From != nil && json.BlockHash != nil {
		setSenderFromServer(json.tx, *json.From, *json.BlockHash)
	}
	return json.tx, err
}

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (ec *Client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	var r *types.Receipt
	err := ec.c.CallContext(ctx, &r, "eth_getTransactionReceipt", txHash)
	if err == nil {
		if r == nil {
			return nil, ethereum.NotFound
		}
	}
	return r, err
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	pending := big.NewInt(-1)
	if number.Cmp(pending) == 0 {
		return "pending"
	}
	return hexutil.EncodeBig(number)
}

type rpcProgress struct {
	StartingBlock hexutil.Uint64
	CurrentBlock  hexutil.Uint64
	HighestBlock  hexutil.Uint64
	PulledStates  hexutil.Uint64
	KnownStates   hexutil.Uint64
}

// SyncProgress retrieves the current progress of the sync algorithm. If there's
// no sync currently running, it returns nil.
func (ec *Client) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	var raw json.RawMessage
	if err := ec.c.CallContext(ctx, &raw, "eth_syncing"); err != nil {
		return nil, err
	}
	// Handle the possible response types
	var syncing bool
	if err := json.Unmarshal(raw, &syncing); err == nil {
		return nil, nil // Not syncing (always false)
	}
	var progress *rpcProgress
	if err := json.Unmarshal(raw, &progress); err != nil {
		return nil, err
	}
	return &ethereum.SyncProgress{
		StartingBlock: uint64(progress.StartingBlock),
		CurrentBlock:  uint64(progress.CurrentBlock),
		HighestBlock:  uint64(progress.HighestBlock),
		PulledStates:  uint64(progress.PulledStates),
		KnownStates:   uint64(progress.KnownStates),
	}, nil
}

// SubscribeNewHead subscribes to notifications about the current blockchain head
// on the given channel.
func (ec *Client) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	return ec.c.EthSubscribe(ctx, ch, "newHeads")
}

// State Access

// NetworkID returns the network ID (also known as the chain ID) for this chain.
func (ec *Client) NetworkID(ctx context.Context) (*big.Int, error) {
	version := new(big.Int)
	var ver string
	if err := ec.c.CallContext(ctx, &ver, "net_version"); err != nil {
		return nil, err
	}
	if _, ok := version.SetString(ver, 10); !ok {
		return nil, fmt.Errorf("invalid net_version result %q", ver)
	}
	return version, nil
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (ec *Client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	var result hexutil.Big
	err := ec.c.CallContext(ctx, &result, "eth_getBalance", account, toBlockNumArg(blockNumber))
	return (*big.Int)(&result), err
}

// StorageAt returns the value of key in the contract storage of the given account.
// The block number can be nil, in which case the value is taken from the latest known block.
func (ec *Client) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	var result hexutil.Bytes
	err := ec.c.CallContext(ctx, &result, "eth_getStorageAt", account, key, toBlockNumArg(blockNumber))
	return result, err
}

// CodeAt returns the contract code of the given account.
// The block number can be nil, in which case the code is taken from the latest known block.
func (ec *Client) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	var result hexutil.Bytes
	err := ec.c.CallContext(ctx, &result, "eth_getCode", account, toBlockNumArg(blockNumber))
	return result, err
}

// NonceAt returns the account nonce of the given account.
// The block number can be nil, in which case the nonce is taken from the latest known block.
func (ec *Client) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	var result hexutil.Uint64
	err := ec.c.CallContext(ctx, &result, "eth_getTransactionCount", account, toBlockNumArg(blockNumber))
	return uint64(result), err
}

// Filters

// FilterLogs executes a filter query.
func (ec *Client) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	var result []types.Log
	arg, err := toFilterArg(q)
	if err != nil {
		return nil, err
	}
	err = ec.c.CallContext(ctx, &result, "eth_getLogs", arg)
	return result, err
}

// SubscribeFilterLogs subscribes to the results of a streaming filter query.
func (ec *Client) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	arg, err := toFilterArg(q)
	if err != nil {
		return nil, err
	}
	return ec.c.EthSubscribe(ctx, ch, "logs", arg)
}

func toFilterArg(q ethereum.FilterQuery) (interface{}, error) {
	arg := map[string]interface{}{
		"address": q.Addresses,
		"topics":  q.Topics,
	}
	if q.BlockHash != nil {
		arg["blockHash"] = *q.BlockHash
		if q.FromBlock != nil || q.ToBlock != nil {
			return nil, fmt.Errorf("cannot specify both BlockHash and FromBlock/ToBlock")
		}
	} else {
		if q.FromBlock == nil {
			arg["fromBlock"] = "0x0"
		} else {
			arg["fromBlock"] = toBlockNumArg(q.FromBlock)
		}
		arg["toBlock"] = toBlockNumArg(q.ToBlock)
	}
	return arg, nil
}

// Pending State

// PendingBalanceAt returns the wei balance of the given account in the pending state.
func (ec *Client) PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	var result hexutil.Big
	err := ec.c.CallContext(ctx, &result, "eth_getBalance", account, "pending")
	return (*big.Int)(&result), err
}

// PendingStorageAt returns the value of key in the contract storage of the given account in the pending state.
func (ec *Client) PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) ([]byte, error) {
	var result hexutil.Bytes
	err := ec.c.CallContext(ctx, &result, "eth_getStorageAt", account, key, "pending")
	return result, err
}

// PendingCodeAt returns the contract code of the given account in the pending state.
func (ec *Client) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	var result hexutil.Bytes
	err := ec.c.CallContext(ctx, &result, "eth_getCode", account, "pending")
	return result, err
}

// PendingNonceAt returns the account nonce of the given account in the pending state.
// This is the nonce that should be used for the next transaction.
func (ec *Client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	var result hexutil.Uint64
	err := ec.c.CallContext(ctx, &result, "eth_getTransactionCount", account, "pending")
	return uint64(result), err
}

// PendingTransactionCount returns the total number of transactions in the pending state.
func (ec *Client) PendingTransactionCount(ctx context.Context) (uint, error) {
	var num hexutil.Uint
	err := ec.c.CallContext(ctx, &num, "eth_getBlockTransactionCountByNumber", "pending")
	return uint(num), err
}

// Contract Calling

// CallContract executes a message call transaction, which is directly executed in the VM
// of the node, but never mined into the blockchain.
//
// blockNumber selects the block height at which the call runs. It can be nil, in which
// case the code is taken from the latest known block. Note that state from very old
// blocks might not be available.
func (ec *Client) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	var hex hexutil.Bytes
	err := ec.c.CallContext(ctx, &hex, "eth_call", toCallArg(msg), toBlockNumArg(blockNumber))
	if err != nil {
		return nil, err
	}
	return hex, nil
}

// PendingCallContract executes a message call transaction using the EVM.
// The state seen by the contract call is the pending state.
func (ec *Client) PendingCallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error) {
	var hex hexutil.Bytes
	err := ec.c.CallContext(ctx, &hex, "eth_call", toCallArg(msg), "pending")
	if err != nil {
		return nil, err
	}
	return hex, nil
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution of a transaction.
func (ec *Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	var hex hexutil.Big
	if err := ec.c.CallContext(ctx, &hex, "eth_gasPrice"); err != nil {
		return nil, err
	}
	return (*big.Int)(&hex), nil
}

// SuggestGasTipCap retrieves the currently suggested gas tip cap after 1559 to
// allow a timely execution of a transaction.
func (ec *Client) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	var hex hexutil.Big
	if err := ec.c.CallContext(ctx, &hex, "eth_maxPriorityFeePerGas"); err != nil {
		return nil, err
	}
	return (*big.Int)(&hex), nil
}

// EstimateGas tries to estimate the gas needed to execute a specific transaction based on
// the current pending state of the backend blockchain. There is no guarantee that this is
// the true gas limit requirement as other transactions may be added or removed by miners,
// but it should provide a basis for setting a reasonable default.
func (ec *Client) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	var hex hexutil.Uint64
	err := ec.c.CallContext(ctx, &hex, "eth_estimateGas", toCallArg(msg))
	if err != nil {
		return 0, err
	}
	return uint64(hex), nil
}

// SendTransaction injects a signed transaction into the pending pool for execution.
//
// If the transaction was a contract creation use the TransactionReceipt method to get the
// contract address after the transaction has been mined.
func (ec *Client) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	data, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	return ec.c.CallContext(ctx, nil, "eth_sendRawTransaction", hexutil.Encode(data))
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (ec *Client) Call(contractAddr common.Address, opts *bind.CallOpts, results *[]interface{}, method, abiStr string, params ...interface{}) error {
	// Don't crash on a lazy user
	if opts == nil {
		opts = new(bind.CallOpts)
	}
	if results == nil {
		results = new([]interface{})
	}
	// cache parsedAbi
	parsedAbi, ok := ec.parsedAbis.Load(contractAddr)
	if !ok {
		p, err := abi.JSON(strings.NewReader(abiStr))
		if err != nil {
			return errors.WithMessagef(err, "parse abi: %s", abiStr)
		}
		parsedAbi = p
		ec.parsedAbis.Store(contractAddr, parsedAbi)
	}
	// Pack the input, call and unpack the results
	input, err := parsedAbi.Pack(method, params...)
	if err != nil {
		return errors.WithMessagef(err, "pack method: %s, params: %+v", method, params)
	}
	var (
		msg    = ethereum.CallMsg{From: opts.From, To: &contractAddr, Data: input}
		ctx    = ensureContext(opts.Context)
		code   []byte
		output []byte
	)
	if opts.Pending {
		output, err = ec.PendingCallContract(ctx, msg)
		if err == nil && len(output) == 0 {
			// Make sure we have a contract to operate on, and bail out otherwise.
			if code, err = ec.PendingCodeAt(ctx, contractAddr); err != nil {
				return err
			} else if len(code) == 0 {
				return bind.ErrNoCode
			}
		}
	} else {
		output, err = ec.CallContract(ctx, msg, opts.BlockNumber)
		if err != nil {
			return err
		}
		if len(output) == 0 {
			// Make sure we have a contract to operate on, and bail out otherwise.
			if code, err = ec.CodeAt(ctx, contractAddr, opts.BlockNumber); err != nil {
				return err
			} else if len(code) == 0 {
				return bind.ErrNoCode
			}
		}
	}

	if len(*results) == 0 {
		res, err := parsedAbi.Unpack(method, output)
		*results = res
		return err
	}
	res := *results
	return parsedAbi.UnpackIntoInterface(res[0], method, output)
}

// BalanceOf query address in contract balance
// returns *big.Int and error
func (ec *Client) BalanceOf(address, contractAddr string) (balance *big.Int, err error) {
	var results = make([]interface{}, 0)
	err = ec.Call(common.HexToAddress(contractAddr), nil, &results, "balanceOf", erc20.ERC20Abi, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	balance, ok := results[0].(*big.Int)
	if !ok {
		return nil, errors.New("results[0] is not *big.Int")
	}
	return
}

// BuildContractTx build contract transaction
func (ec *Client) BuildContractTx(privKey, method, abiStr string, contract *common.Address, opts *bind.TransactOpts, params ...interface{}) (tx *types.Transaction, err error) {
	if contract == nil {
		return nil, errors.New("contract is nil")
	}
	// decode private key
	pKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		err = errors.WithMessage(err, "hex private key to ECDSA key: ")
		return
	}
	from := crypto.PubkeyToAddress(pKey.PublicKey)
	// Don't crash on a lazy user
	if opts == nil {
		opts = &bind.TransactOpts{From: from}
	}

	// pack input params and cache parsedAbi
	parsedAbi, ok := ec.parsedAbis.Load(*contract)
	if !ok {
		p, err := abi.JSON(strings.NewReader(abiStr))
		if err != nil {
			return nil, errors.WithMessagef(err, "parse abi: %s", abiStr)
		}
		parsedAbi = p
		ec.parsedAbis.Store(*contract, parsedAbi)
	}

	// Pack the input, call and unpack the results
	input, err := parsedAbi.Pack(method, params...)
	if err != nil {
		err = errors.WithMessagef(err, "pack method: %s, params: %+v", method, params)
		return
	}

	// Ensure a valid value field and resolve the account nonce
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	var nonce uint64
	if opts.Nonce == nil {
		nonce, err = ec.PendingNonceAt(ensureContext(opts.Context), from)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
		}
	} else {
		nonce = opts.Nonce.Uint64()
	}

	// Figure out reasonable gas price values
	if opts.GasPrice != nil && (opts.GasFeeCap != nil || opts.GasTipCap != nil) {
		return nil, errors.New("both gasPrice and (maxFeePerGas or maxPriorityFeePerGas) specified")
	}
	head, err := ec.HeaderByNumber(ensureContext(opts.Context), nil)
	if err != nil {
		return nil, errors.WithMessage(err, "header by number")
	}
	if head.BaseFee != nil && opts.GasPrice == nil {
		if opts.GasTipCap == nil {
			tip, err := ec.SuggestGasTipCap(ensureContext(opts.Context))
			if err != nil {
				return nil, err
			}
			opts.GasTipCap = tip
		}
		if opts.GasFeeCap == nil {
			gasFeeCap := new(big.Int).Add(
				opts.GasTipCap,
				new(big.Int).Mul(head.BaseFee, big.NewInt(2)),
			)
			opts.GasFeeCap = gasFeeCap
		}
		if opts.GasFeeCap.Cmp(opts.GasTipCap) < 0 {
			return nil, fmt.Errorf("maxFeePerGas (%v) < maxPriorityFeePerGas (%v)", opts.GasFeeCap, opts.GasTipCap)
		}
	} else {
		if opts.GasFeeCap != nil || opts.GasTipCap != nil {
			return nil, errors.New("maxFeePerGas or maxPriorityFeePerGas specified but london is not active yet")
		}
		if opts.GasPrice == nil {
			price, err := ec.SuggestGasPrice(ensureContext(opts.Context))
			if err != nil {
				return nil, err
			}
			opts.GasPrice = price
		}
	}

	gasLimit := opts.GasLimit
	if gasLimit == 0 {
		// Gas estimation cannot succeed without code for method invocations
		if code, err := ec.PendingCodeAt(ensureContext(opts.Context), *contract); err != nil {
			return nil, err
		} else if len(code) == 0 {
			return nil, bind.ErrNoCode
		}
		// If the contract surely has code (or code is not needed), estimate the transaction
		msg := ethereum.CallMsg{From: opts.From, To: contract, GasPrice: opts.GasPrice, GasTipCap: opts.GasTipCap, GasFeeCap: opts.GasFeeCap, Value: value, Data: input}
		gasLimit, err = ec.EstimateGas(ensureContext(opts.Context), msg)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
		}
	}

	// Create the transaction, sign it and schedule it for execution
	var rawTx *types.Transaction
	if opts.GasFeeCap == nil {
		baseTx := &types.LegacyTx{
			Nonce:    nonce,
			GasPrice: opts.GasPrice,
			To:       contract,
			Gas:      gasLimit,
			Value:    value,
			Data:     input,
		}
		rawTx = types.NewTx(baseTx)
	} else {
		baseTx := &types.DynamicFeeTx{
			Nonce:     nonce,
			GasFeeCap: opts.GasFeeCap,
			GasTipCap: opts.GasTipCap,
			Gas:       gasLimit,
			Value:     value,
			To:        contract,
			Data:      input,
		}
		rawTx = types.NewTx(baseTx)
	}

	if ec.chainID == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(ec.timeout))
		chainID, err := ec.ChainID(ctx)
		cancel()
		if err != nil {
			return nil, errors.WithMessage(err, "get chain id: ")
		}
		ec.chainID = chainID
	}

	signedTx, err := types.SignTx(rawTx, types.NewLondonSigner(ec.chainID), pKey)
	if err != nil {
		err = errors.WithMessage(err, "signed raw tx")
		return
	}
	return signedTx, nil
}

func (ec *Client) BuildTransferTx(privKey, to string, opts *bind.TransactOpts) (tx *types.Transaction, err error) {
	// decode private key
	pKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		err = errors.WithMessage(err, "hex private key to ECDSA key: ")
		return
	}
	from := crypto.PubkeyToAddress(pKey.PublicKey)

	// Don't crash on a lazy user
	if opts == nil {
		opts = &bind.TransactOpts{From: from, GasLimit: 21000}
	}

	// Ensure a valid value field and resolve the account nonce
	value := opts.Value
	if value == nil {
		opts.Value = new(big.Int)
	}
	var nonce uint64
	if opts.Nonce == nil {
		nonce, err = ec.PendingNonceAt(ensureContext(opts.Context), from)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve account nonce: %v", err)
		}
	} else {
		nonce = opts.Nonce.Uint64()
	}

	// Figure out reasonable gas price values
	if opts.GasPrice != nil && (opts.GasFeeCap != nil || opts.GasTipCap != nil) {
		return nil, errors.New("both gasPrice and (maxFeePerGas or maxPriorityFeePerGas) specified")
	}
	head, err := ec.HeaderByNumber(ensureContext(opts.Context), nil)
	if err != nil {
		return nil, errors.WithMessage(err, "header by number")
	}
	if head.BaseFee != nil && opts.GasPrice == nil {
		if opts.GasTipCap == nil {
			tip, err := ec.SuggestGasTipCap(ensureContext(opts.Context))
			if err != nil {
				return nil, err
			}
			opts.GasTipCap = tip
		}
		if opts.GasFeeCap == nil {
			gasFeeCap := new(big.Int).Add(
				opts.GasTipCap,
				new(big.Int).Mul(head.BaseFee, big.NewInt(2)),
			)
			opts.GasFeeCap = gasFeeCap
		}
		if opts.GasFeeCap.Cmp(opts.GasTipCap) < 0 {
			return nil, fmt.Errorf("maxFeePerGas (%v) < maxPriorityFeePerGas (%v)", opts.GasFeeCap, opts.GasTipCap)
		}
	} else {
		if opts.GasFeeCap != nil || opts.GasTipCap != nil {
			return nil, errors.New("maxFeePerGas or maxPriorityFeePerGas specified but london is not active yet")
		}
		if opts.GasPrice == nil {
			price, err := ec.SuggestGasPrice(ensureContext(opts.Context))
			if err != nil {
				return nil, err
			}
			opts.GasPrice = price
		}
	}
	toAddr := common.HexToAddress(to)
	// Create the transaction, sign it and schedule it for execution
	var rawTx *types.Transaction
	if opts.GasFeeCap == nil {
		baseTx := types.LegacyTx{
			Nonce:    nonce,
			To:       &toAddr,
			GasPrice: opts.GasPrice,
			Gas:      opts.GasLimit,
			Value:    opts.Value,
			Data:     []byte{},
		}
		rawTx = types.NewTx(&baseTx)
	} else {
		baseTx := types.DynamicFeeTx{
			Nonce:     nonce,
			GasFeeCap: opts.GasFeeCap,
			GasTipCap: opts.GasTipCap,
			Gas:       opts.GasLimit,
			Value:     value,
			To:        &toAddr,
			Data:      []byte{},
		}
		rawTx = types.NewTx(&baseTx)
	}

	if ec.chainID == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(ec.timeout))
		chainID, err := ec.ChainID(ctx)
		cancel()
		if err != nil {
			return nil, errors.WithMessage(err, "get chain id: ")
		}
		ec.chainID = chainID
	}

	signedTx, err := types.SignTx(rawTx, types.NewLondonSigner(ec.chainID), pKey)
	if err != nil {
		err = errors.WithMessage(err, "signed raw tx")
		return
	}
	return signedTx, nil
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}

// ensureContext is a helper method to ensure a context is not nil, even if the
// user specified it as such.
func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
