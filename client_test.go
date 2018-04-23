package ethfw

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/xlab/closer"

	"github.com/AtlantPlatform/ethfw/sol"
)

// workflow imitation
var path = `/var/chain/geth.ipc`
var contrAddress = common.HexToAddress("0x22714a7e5ff13df0591d821fa760f88a5b0e60de")

// any transaction address
var txHash = common.HexToHash("0xa5e5dee32ed9aa6084e5a48ddb8efd872cd1b827f71e4475332b4739966d1775")

// any password
var password = "123456"

// ATL contract
var sampleABI = []byte(`[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_holder","type":"address"},{"name":"_value","type":"uint256"}],"name":"mint","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"ico","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"unfreeze","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"tokensAreFrozen","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"remaining","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"inputs":[{"name":"_ico","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`)
var sampleBIN = `0x606060405260408051908101604052600c81527f41544c414e5420546f6b656e00000000000000000000000000000000000000006020820152600390805161004b929160200190610106565b5060408051908101604052600381527f41544c000000000000000000000000000000000000000000000000000000000060208201526004908051610093929160200190610106565b5060126005556006805460a060020a60ff0219167401000000000000000000000000000000000000000017905534156100cb57600080fd5b604051602080610a328339810160405280805160068054600160a060020a031916600160a060020a0392909216919091179055506101a19050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061014757805160ff1916838001178555610174565b82800160010185558215610174579182015b82811115610174578251825591602001919060010190610159565b50610180929150610184565b5090565b61019e91905b80821115610180576000815560010161018a565b90565b610882806101b06000396000f3006060604052600436106100c45763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde0381146100c9578063095ea7b31461015357806318160ddd1461017757806323b872dd1461019c578063313ce567146101c457806340c10f19146101d75780635d452201146101f95780636a28f0001461022857806370a082311461023b57806395d89b411461025a578063a9059cbb1461026d578063ca67065f1461028f578063dd62ed3e146102b6575b600080fd5b34156100d457600080fd5b6100dc6102db565b60405160208082528190810183818151815260200191508051906020019080838360005b83811015610118578082015183820152602001610100565b50505050905090810190601f1680156101455780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b341561015e57600080fd5b610175600160a060020a0360043516602435610379565b005b341561018257600080fd5b61018a6103af565b60405190815260200160405180910390f35b34156101a757600080fd5b610175600160a060020a03600435811690602435166044356103b5565b34156101cf57600080fd5b61018a6103ed565b34156101e257600080fd5b610175600160a060020a03600435166024356103f3565b341561020457600080fd5b61020c610493565b604051600160a060020a03909116815260200160405180910390f35b341561023357600080fd5b6101756104a2565b341561024657600080fd5b61018a600160a060020a03600435166104dd565b341561026557600080fd5b6100dc6104f8565b341561027857600080fd5b610175600160a060020a0360043516602435610563565b341561029a57600080fd5b6102a2610595565b604051901515815260200160405180910390f35b34156102c157600080fd5b61018a600160a060020a03600435811690602435166105b6565b60038054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156103715780601f1061034657610100808354040283529160200191610371565b820191906000526020600020905b81548152906001019060200180831161035457829003601f168201915b505050505081565b60065474010000000000000000000000000000000000000000900460ff16156103a157600080fd5b6103ab82826105e1565b5050565b60005481565b60065474010000000000000000000000000000000000000000900460ff16156103dd57600080fd5b6103e8838383610645565b505050565b60055481565b60065433600160a060020a0390811691161461040e57600080fd5b80151561041a57600080fd5b6000546a7c13bc4b2c133c56000000908201111561043757600080fd5b600160a060020a0382166000818152600160205260408082208054850190558154840182557fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9084905190815260200160405180910390a35050565b600654600160a060020a031681565b60065433600160a060020a039081169116146104bd57600080fd5b6006805474ff000000000000000000000000000000000000000019169055565b600160a060020a031660009081526001602052604090205490565b60048054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156103715780601f1061034657610100808354040283529160200191610371565b60065474010000000000000000000000000000000000000000900460ff161561058b57600080fd5b6103ab8282610750565b60065474010000000000000000000000000000000000000000900460ff1681565b600160a060020a03918216600090815260026020908152604080832093909416825291909152205490565b600160a060020a03338116600081815260026020908152604080832094871680845294909152908190208490557f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9259084905190815260200160405180910390a35050565b600160a060020a038084166000908152600260209081526040808320338516845282528083205493861683526001909152902054610689908363ffffffff61081b16565b600160a060020a0380851660009081526001602052604080822093909355908616815220546106be908363ffffffff61083316565b600160a060020a0385166000908152600160205260409020556106e7818363ffffffff61083316565b600160a060020a03808616600081815260026020908152604080832033861684529091529081902093909355908516917fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9085905190815260200160405180910390a350505050565b6040604436101561076057600080fd5b600160a060020a033316600090815260016020526040902054610789908363ffffffff61083316565b600160a060020a0333811660009081526001602052604080822093909355908516815220546107be908363ffffffff61081b16565b600160a060020a0380851660008181526001602052604090819020939093559133909116907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9085905190815260200160405180910390a3505050565b600082820161082c84821015610847565b9392505050565b600061084183831115610847565b50900390565b80151561085357600080fd5b505600a165627a7a72305820af9feb4d3906db4ebd65e1c60f35944c9a78c462c716e42b3d1b1d1b2ca1ea120029`

var cli Client
var accountFirst common.Address
var accountSecond common.Address
var opts *bind.TransactOpts
var boundContract *BoundContract

func TestMain(m *testing.M) {
	setUpClient(getIPCPath())
	code := m.Run()
	shutDownClient()
	os.Exit(code)

}
func getIPCPath() string {
	_, b, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("failed to parse current directory")
	}
	return filepath.Dir(b) + path
}

func setUpClient(socket string) {
	rpcCli, err := rpc.Dial(socket)
	if err != nil {
		log.Fatalf("failed to dial socket %v: %v", socket, err)
	}
	closer.Bind(func() {
		rpcCli.Close()
	})
	cli = NewClient(rpcCli)
}

func shutDownClient() {
	closer.Close()
}

func TestAccounts(t *testing.T) {
	accs, err := cli.PersonalAccounts(context.Background())
	if err != nil {
		t.Errorf("failed to retreive accounts %v: %v", accs, err)
		return
	}
	fmt.Println("TestAccounts:")
	for _, acc := range accs {
		fmt.Println(acc.String())
	}
	accountFirst = accs[0]
	if len(accs) > 1 {
		accountSecond = accs[1]
	}
}

func TestBalanceAt(t *testing.T) {
	bal, err := cli.BalanceAt(context.Background(), accountFirst, nil)
	if err != nil {
		t.Errorf("failed to retreive balance %v: %v", bal, err)
		return
	}
	fmt.Println("TestBalanceAt:", bal)
}

func TestCodeAt(t *testing.T) {
	code, err := cli.CodeAt(context.Background(), accountFirst, nil)
	if err != nil {
		t.Errorf("failed to retreive code %v: %v", code, err)
		return
	}
	fmt.Println("TestCodeAt:", code)
}

func TestPendingCodeAt(t *testing.T) {
	pCode, err := cli.PendingCodeAt(context.Background(), accountFirst)
	if err != nil {
		t.Errorf("failed to retreive pending code %v: %v", pCode, err)
		return
	}
	fmt.Println("TestPendingCodeAt:", pCode)
}

func TestPendingNonceAt(t *testing.T) {
	pNonce, err := cli.PendingNonceAt(context.Background(), accountFirst)
	if err != nil {
		t.Errorf("failed to retreive pending nonce %v: %v", pNonce, err)
		return
	}
	fmt.Println("TestPendingNonceAt:", pNonce)
}

func TestSuggestGasPrice(t *testing.T) {
	price, err := cli.SuggestGasPrice(context.Background())
	if err != nil {
		t.Errorf("failed to retreive gas price %v: %v", price, err)
		return
	}
	fmt.Println("TestSuggestGasPrice:", price)
}

func TestEstimateGas(t *testing.T) {
	msg := ethereum.CallMsg{From: accountFirst, To: &accountSecond}
	eGas, err := cli.EstimateGas(context.Background(), msg)
	if err != nil {
		t.Errorf("failed to retreive estimate gas %v: %v", eGas, err)
		return
	}
	fmt.Println("TestEstimateGas:", eGas)
}

func TestFilterLogs(t *testing.T) {
	filter := ethereum.FilterQuery{}
	logs, err := cli.FilterLogs(context.Background(), filter)
	if err != nil {
		t.Errorf("failed to retreive logs %v: %v", logs, err)
		return
	}
	fmt.Println("TestFilterLogs:", logs)
}

func TestSubscribeFilterLogs(t *testing.T) {
	logs := make(chan types.Log, 128)
	defer close(logs)
	filter := ethereum.FilterQuery{}
	sLogs, err := cli.SubscribeFilterLogs(context.Background(), filter, logs)
	if err != nil {
		t.Errorf("failed to subscribe logs %v: %v", sLogs, err)
		return
	}
	fmt.Println("TestSubscribeFilterLogs:", sLogs)

}

func TestTransactionReceipt(t *testing.T) {
	receipt, err := cli.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		t.Errorf("failed to retreive transaction receipt %v: %v", receipt, err)
		return
	}
	fmt.Println("TestTransactionReceipt:", receipt.TxHash.String())
}

func TestTransaction(t *testing.T) {
	tx, isPending, err := cli.Transaction(context.Background(), txHash)
	if err != nil {
		t.Errorf("failed to get transaction hash: %v", err)
		return
	}
	fmt.Printf("TestTransaction: %s, %t\n", tx.Hash().String(), isPending)

}

func TestBindContract(t *testing.T) {
	//must be fail
	var err error
	contract := sol.Contract{Name: "test", Address: contrAddress, ABI: sampleABI, Bin: sampleBIN}
	boundContract, err = cli.BindContract(&contract)
	if err != nil {
		t.Errorf("failed to bind contract %#v: %v", boundContract, err)
		return
	}
	fmt.Printf("TestBindContract: %#v\n", boundContract.abi.Methods["name"])
}

func TestOpts(t *testing.T) {
	var err error
	opts, err = cli.TransactOpts(context.Background(), accountFirst, password)
	if err != nil {
		t.Errorf("failed to assemble transaction options: %v", err)
		return
	}
	fmt.Printf("TestOpts: %#v\n", opts)

}

func TestTransact(t *testing.T) {
	if opts == nil {
		t.Error("no opts were configured")
		return
	}
	tx, err := boundContract.Transact(opts, "balanceOf", opts.From)
	if err != nil {
		t.Errorf("failed to transact %#v: %v", tx, err)
		return
	}
	fmt.Printf("TestTransact: %#v\n", tx)
}

//fallback method
func TestTransfer(t *testing.T) {
	if opts == nil {
		t.Error("no opts were configured")
		return
	}
	tx, err := boundContract.Transfer(opts)
	if err != nil {
		t.Errorf("failed to trasfer %#v: %v", tx, err)
		return
	}
	fmt.Printf("TestTransfer: %#v\n", tx)
}

func TestDeploy(t *testing.T) {
	if opts == nil {
		t.Error("no opts were configured")
		return
	}
	addr, tx, err := boundContract.Deploy(opts)
	if err != nil {
		t.Errorf("failed to deploy: %v", err)
		return
	}
	fmt.Printf("TestDeploy: %v %v\n", tx, addr)
}
