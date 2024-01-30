package util

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/yapkah/go-api/pkg/float"
)

// BaseDecimal decimal for base token
const BaseDecimal int64 = 18

// SendERC20 func
// sending ERC20 token, can be used for LIGA / PAC token which follow ERC20 format
func SendERC20(chainID int64, privateKey, to, contractAddress string, nonce, maxGas uint64, value float64, decimal int64) (string, error) {

	var baseValue float64

	// method id
	methodID, _ := hexutil.Decode("0xa9059cbb")

	// to address
	toAddress := common.HexToAddress(to)
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)

	// transfer token value
	tokenValue := float.ValueConverting(value, decimal)
	paddedValue := common.LeftPadBytes(tokenValue.Bytes(), 32)

	// function data
	// transfer(address _to, uint256 _value)
	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedValue...)

	// sign transaction
	return SignTransaction(chainID, privateKey, contractAddress, nonce, maxGas, baseValue, data)
}

// SignTransaction sign eth transaction
func SignTransaction(chainID int64, privateKey, to string, nonce, maxGas uint64, value float64, data []byte) (string, error) {
	pk, err := crypto.HexToECDSA(strings.ReplaceAll(privateKey, "0x", ""))
	if err != nil {
		return "", err
	}

	// get from address
	toAddress := common.HexToAddress(to)
	gasPrice := big.NewInt(0)
	valueWei := float.ValueConverting(value, BaseDecimal)

	tx := types.NewTransaction(nonce, toAddress, valueWei, maxGas, gasPrice, data)

	chainIDBigInt := big.NewInt(chainID)
	signer := types.NewEIP155Signer(chainIDBigInt)
	signedTx, err := types.SignTx(tx, signer, pk)
	if err != nil {
		return "", err
	}

	var buff bytes.Buffer
	signedTx.EncodeRLP(&buff)
	rawTransaction := fmt.Sprintf("0x%x", buff.Bytes())

	return rawTransaction, nil
}

// DecodeSigningKey
func DecodeSigningKey(rawTx string) (map[string]interface{}, error) {

	tx := new(types.Transaction)
	rawTxBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		return nil, err
	}

	rlp.DecodeBytes(rawTxBytes, &tx)

	return nil, err
	// msg, err := tx.AsMessage(types.NewEIP155Signer(tx.ChainId()))
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("msg: ", msg)
	// arrDataReturn := map[string]interface{}{
	// 	"From":      msg.From(),
	// 	"To":        msg.To(),
	// 	"Gas":       msg.Gas(),
	// 	"GasPrice":  msg.GasPrice(),
	// 	"Value":     msg.Value(),
	// 	"From-Hash": msg.From().Hash(),
	// 	"From-Hex":  msg.From().Hex(),
	// }

	// return arrDataReturn, nil
}
