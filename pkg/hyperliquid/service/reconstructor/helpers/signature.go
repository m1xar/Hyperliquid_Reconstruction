package helpers

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func VerifySignature(expectedAddress, signatureHex, message string) bool {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	msgHash := crypto.Keccak256Hash([]byte(msg))

	sig := common.FromHex(signatureHex)
	if len(sig) != 65 {
		return false
	}
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	if sig[64] != 0 && sig[64] != 1 {
		return false
	}

	pubKey, err := crypto.SigToPub(msgHash.Bytes(), sig)
	if err != nil {
		return false
	}

	recovered := crypto.PubkeyToAddress(*pubKey).Hex()
	return strings.EqualFold(recovered, expectedAddress)
}
