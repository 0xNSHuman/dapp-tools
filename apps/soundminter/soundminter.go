package soundminter

import (
	"bytes"
	"context"
	"encoding/hex"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/0xNSHuman/dapp-tools/client"
	"github.com/0xNSHuman/dapp-tools/http"
	"github.com/0xNSHuman/dapp-tools/utils"
	"github.com/0xNSHuman/dapp-tools/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type EventField_MerkleDropMintCreated uint16

const (
	merkleDropMinterAddressHex = "0xeae422887230c0ffb91fd8f708f5fdd354c92f2f"
)

const (
	EventField_MerkleDropMintCreated_EventName EventField_MerkleDropMintCreated = iota
	EventField_MerkleDropMintCreated_EditionAddress
	EventField_MerkleDropMintCreated_MintId
	EventField_MerkleDropMintCreated_MerkleRootHash
	EventField_MerkleDropMintCreated_Price
	EventField_MerkleDropMintCreated_StartTime
	EventField_MerkleDropMintCreated_EndTime
	EventField_MerkleDropMintCreated_AffiliateFeeBPS
	EventField_MerkleDropMintCreated_MaxMintable
	EventField_MerkleDropMintCreated_MaxMintablePerAccount
)

type MerkleProofResponse struct {
	UnhashedLeaf string   `json:"unhashedLeaf"`
	Proof        []string `json:"proof"`
}

type Soundminter struct {
	editionAddress common.Address
	abi            abi.ABI
	client         *client.Client
	wallet         *wallet.WalletKeeper
}

func NewSoundminter(
	rpcEndpoint string,
	walletUI wallet.WalletUI,
	editionAddress common.Address,
) (*Soundminter, error) {
	client, err := client.NewClient(rpcEndpoint)
	if err != nil {
		return nil, err
	}

	wallet, err := wallet.NewWalletKeeper(walletUI)
	if err != nil {
		return nil, err
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	abiPath := filepath.Join(
		userHomeDir,
		"evm", "apps", "soundminter", "abi",
		"SoundMerkleDropMinter.json",
	)
	contractABIFile, err := os.ReadFile(abiPath)
	if err != nil {
		return nil, err
	}

	abi, err := abi.JSON(bytes.NewReader(contractABIFile))
	if err != nil {
		return nil, err
	}

	return &Soundminter{
		editionAddress: editionAddress,
		abi:            abi,
		client:         client,
		wallet:         wallet,
	}, nil
}

func (sm *Soundminter) Automint() (*string, error) {
	// Step 1:
	//		Source mint details from the edition contract deployment logs:
	//			- mintId
	//			- mint price
	//			- merkle root hash
	//			- start time (as UNIX timestamp, not block num)

	blockNum, err := sm.client.EthClient.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}

	logs, err := sm.client.ReadLogs(
		big.NewInt(int64(blockNum-216_000)), // Last 30 days lookup
		big.NewInt(int64(blockNum)),
		sm.abi,
		common.HexToAddress(merkleDropMinterAddressHex),
		[][]string{
			// Topic 0 targets
			{
				"MerkleDropMintCreated(address,uint128,bytes32,uint96,uint32,uint32,uint16,uint32,uint32)",
			},
			// Topic 1 targets
			{
				string(sm.editionAddress.Hex()),
			},
		},
	)
	if err != nil {
		return nil, err
	}

	// Trim leading zeros but leave the '0x' prefix
	mintIdHex := strings.TrimLeft(logs[EventField_MerkleDropMintCreated_MintId].(string), "0x")
	mintIdHex = strings.Join([]string{"0x", mintIdHex}, "")
	mintId, err := hexutil.DecodeBig(mintIdHex)
	if err != nil {
		return nil, err
	}

	merkleRootHashBytes32 := logs[EventField_MerkleDropMintCreated_MerkleRootHash].([32]uint8)
	merkleRootHash := common.Bytes2Hex(merkleRootHashBytes32[:])
	_ = merkleRootHash

	price := logs[EventField_MerkleDropMintCreated_Price].(*big.Int)

	startTimestamp := logs[EventField_MerkleDropMintCreated_StartTime].(uint32)
	_ = startTimestamp

	// Step 2:
	//		Get the merkle proof from for the currently stored wallet
	// 		using the off-chain service where it's stored

	pubkey, err := sm.wallet.PublicKey()
	if err != nil {
		return nil, err
	}

	var proofResponse MerkleProofResponse
	query := strings.Join(
		[]string{
			"https://lanyard.org/api/v1/proof",
			"?root=", merkleRootHash,
			"&unhashedLeaf=", pubkey,
		},
		"",
	)
	err = http.GetObject(query, &proofResponse)
	if err != nil {
		return nil, err
	}

	merkleProofBytes := make([][32]byte, len(proofResponse.Proof))

	for i, hash := range proofResponse.Proof {
		bytes, _ := hex.DecodeString(strings.Trim(hash, "0x"))
		bytes32 := new([32]byte)
		copy(bytes32[:], bytes)
		merkleProofBytes[i] = *bytes32
	}

	// Step 3:
	//		Fire the mint TX

	return sm.mint(
		sm.editionAddress,
		mintId,
		uint32(1),
		merkleProofBytes,
		price,
	)
}

// The contract mint function invoker.
//
// Params:
//
//	editionAddress 		The address of the deployed Sound.xyz edition contract.
//	mintId 				The ID generated during the deployment. See `MintConfigCreated`
//						log event emitted in the deployment transaction.
//	requestedQuantity	How many tokens to mint.
//	merkleProof			Merkle proof generated earlier. The Merkle tree must be sourced
//						from somewhere for that (like an allowlist).
//	price				Mint price.
func (sm *Soundminter) mint(
	editionAddress common.Address,
	mintId *big.Int,
	requestedQuantity uint32,
	merkleProof [][32]byte,
	price *big.Int,
) (*string, error) {
	pubkey, err := sm.wallet.PublicKey()
	if err != nil {
		return nil, err
	}

	from := common.HexToAddress(pubkey)
	to := common.HexToAddress(merkleDropMinterAddressHex)
	value := price

	callData, err := sm.abi.Pack(
		"mint",
		editionAddress,
		mintId,
		requestedQuantity,
		merkleProof,
		common.Address{}, // affiliate address
	)
	if err != nil {
		return nil, err
	}

	tx, err := utils.EncodeTransaction(sm.client, from, to, value, callData)
	if err != nil {
		return nil, err
	}

	chainId, err := sm.client.ChainID()
	if err != nil {
		return nil, err
	}

	signedTx, err := sm.wallet.SignTransaction(chainId, tx)
	if err != nil {
		return nil, err
	}

	txHash, err := sm.client.SendTransaction(signedTx)
	if err != nil {
		return nil, err
	}

	return txHash, nil
}
