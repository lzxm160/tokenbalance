package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	// "github.com/mkideal/cli"
	// "log"
	"math/big"
	"os"
)

var conn *ethclient.Client

var GethLocation string
var UsePort string
var UseIP string
var version string = "v0.0.1"

var decimals uint8

func main() {
	ConnectGeth()

	name, tokenCorrected, symbol, tokenDecimals, ethCorrected, maxBlock, err := GetAccount("0x86fa049857e0209aa7d9e616f7eb3b3b78ecfdb0", "0x4752218e54De423F86c0501933917aea08c8FED5")
	if err != nil {
		fmt.Println("getaccount:", err)
		return
	}
	fmt.Println(name, ":", tokenCorrected, ":", symbol, ":", tokenDecimals, ":", ethCorrected, ":", maxBlock)
}

func ConnectGeth() {
	var err error
	conn, err = ethclient.Dial("/usr/local/geth/data/geth.ipc")
	if err != nil {
		fmt.Println("Failed to connect to the Ethereum client: %v", err)
	} else {
		fmt.Println("Connected to Geth at: /usr/local/geth/data/geth.ipc")
	}
}

func GetAccount(contract string, wallet string) (string, string, string, uint8, string, uint64, error) {
	var err error
	var symbol string

	token, err := NewTokenCaller(common.HexToAddress(contract), conn)
	if err != nil {
		fmt.Println("Failed to instantiate a Token contract: %v", err)
		return "error", "0.0", "error", 0, "0.0", 0, err
	}

	getBlock, err := conn.BlockByNumber(context.Background(), nil)
	if err != nil {
		fmt.Println("Failed to get current block number: ", err)
		return "error", "0.0", "error", 0, "0.0", 0, err
	}

	maxBlock := getBlock.NumberU64()

	address := common.HexToAddress(wallet)
	if err != nil {
		fmt.Println("Failed hex address: "+wallet, err)
		return "error", "0.0", "error", 0, "0.0", 0, err
	}

	ethAmount, err := conn.BalanceAt(context.Background(), address, nil)
	if err != nil {
		fmt.Println("Failed to get ethereum balance from address: ", address, err)
		return "error", "0.0", "error", 0, "0.0", 0, err
	}

	balance, err := token.BalanceOf(nil, address)
	if err != nil {
		fmt.Println("Failed to get balance from contract: "+contract, err)
		return "error", "0.0", "error", 0, "0.0", 0, err
	}

	// the popular coin EOS doesn't have a symbol
	if common.HexToAddress(contract) == common.HexToAddress("0x86fa049857e0209aa7d9e616f7eb3b3b78ecfdb0") {
		symbol = "EOS"
	} else {
		symbol, err = token.Symbol(nil)
		if err != nil {
			fmt.Println("Failed to get symbol from contract: "+contract, err)
			return "error", "0.0", "error", 0, "0.0", 0, err
		}
	}
	tokenDecimals, err := token.Decimals(nil)
	if err != nil {
		fmt.Println("Failed to get decimals from contract: "+contract, err)
		return "error", "0.0", "error", 0, "0.0", 0, err
	}
	name, err := token.Name(nil)
	if err != nil {
		fmt.Println("Failed to retrieve token name from contract: "+contract, err)
		return "error", "0.0", "error", 0, "0.0", 0, err
	}

	ethCorrected := BigIntDecimal(ethAmount, 18)
	tokenCorrected := BigIntDecimal(balance, int(tokenDecimals))

	return name, tokenCorrected, symbol, tokenDecimals, ethCorrected, maxBlock, err

}

func BigIntDecimal(balance *big.Int, decimals int) string {
	if balance.String() == "0" {
		return "0"
	}
	var newNum string
	for k, v := range balance.String() {
		if k == len(balance.String())-decimals {
			newNum += "."
		}
		newNum += string(v)
	}
	stringBytes := bytes.TrimRight([]byte(newNum), "0")
	newNum = string(stringBytes)
	if stringBytes[len(stringBytes)-1] == 46 {
		newNum += "0"
	}
	if stringBytes[0] == 46 {
		newNum = "0" + newNum
	}
	return newNum
}
