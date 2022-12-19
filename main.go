package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/0xNSHuman/dapp-tools/apps"
	"github.com/0xNSHuman/dapp-tools/apps/scoretable"
	"github.com/0xNSHuman/dapp-tools/apps/soundminter"
	"github.com/0xNSHuman/dapp-tools/client"
	"github.com/0xNSHuman/dapp-tools/ui"
	"github.com/0xNSHuman/dapp-tools/wallet"
	"github.com/ethereum/go-ethereum/common"
)

var (
	walletCLI   *ui.WalletCLI
	rpcEndpoint string
)

// +++++++++++++++++++++++++++++++++++++++
// 		  		    ENTRY
// +++++++++++++++++++++++++++++++++++++++

func main() {
	walletCLI = ui.NewWalletCLI()

	flag.StringVar(&rpcEndpoint, "rpc", "", "rpc endpoint")

	flag.Parse()
	args := flag.Args()
	handleArgs(args)
}

func handleArgs(args []string) {
	if len(args) == 0 {
		fmt.Println("Command not provided")
		return
	}

	switch args[0] {
	case "wallet":
		handleWalletCommand(args[1:])
	case "block":
		handleBlockCommand(args[1:])
	case "app":
		handleAppCommand(args[1:])
	default:
		fmt.Println("Command not found")
		return
	}
}

// +++++++++++++++++++++++++++++++++++++++
// 		  	   WALLET COMMANDS
// +++++++++++++++++++++++++++++++++++++++

func handleWalletCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	switch args[0] {
	case "create":
		executeWalletCreateCommand(args[1:])
	case "pubkey":
		executeWalletPubkeyCommand(args[1:])
	case "export":
		executeWalletExportCommand(args[1:])
	case "import":
		executeWalletImportCommand(args[1:])
	case "balance":
		executeWalletBalanceCommand(args[1:])
	default:
		fmt.Println("Invalid command usage")
		return
	}
}

func executeWalletCreateCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	passphrase := findArg(args, "--pwd=")

	if passphrase == nil {
		fmt.Println("--pwd parameter is required")
		return
	}

	walletKeeper, err := wallet.NewWalletKeeper(walletCLI)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = walletKeeper.CreateWallet(*passphrase)
	if err != nil {
		fmt.Println(err)
		return
	}

	pubkey, err := walletKeeper.PublicKey()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Wallet created! Public address:", pubkey)
}

func executeWalletPubkeyCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	passphrase := findArg(args, "--pwd=")

	if passphrase == nil {
		fmt.Println("--pwd parameter is required")
		return
	}

	walletKeeper, err := wallet.NewWalletKeeper(walletCLI)
	if err != nil {
		fmt.Println(err)
		return
	}

	pubkey, err := walletKeeper.PublicKey()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Public address:", pubkey)
}

func executeWalletExportCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	passphrase := findArg(args, "--pwd=")

	if passphrase == nil {
		fmt.Println("--pwd parameter is required")
		return
	}

	walletKeeper, err := wallet.NewWalletKeeper(walletCLI)
	if err != nil {
		fmt.Println(err)
		return
	}

	privateKey, err := walletKeeper.ExportWallet(wallet.ExportModePrivateKey, *passphrase)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Private key:", common.Bytes2Hex(privateKey))
}

func executeWalletImportCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	passphrase := findArg(args, "--pwd=")
	privateKey := findArg(args, "--privateKey=")

	if passphrase == nil {
		fmt.Println("--pwd parameter is required")
		return
	}
	if privateKey == nil {
		fmt.Println("--privateKey parameter is required")
		return
	}

	walletKeeper, err := wallet.NewWalletKeeper(walletCLI)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = walletKeeper.ImportWallet(wallet.ImportModePrivateKey, common.Hex2Bytes(*privateKey), *passphrase)
	if err != nil {
		fmt.Println(err)
		return
	}

	pubkey, err := walletKeeper.PublicKey()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Wallet imported! Public address:", pubkey)
}

func executeWalletBalanceCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	rpcEndpoint := findArg(args, "--rpc=")
	passphrase := findArg(args, "--pwd=")

	if passphrase == nil {
		fmt.Println("--pwd parameter is required")
		return
	}
	if rpcEndpoint == nil {
		fmt.Println("--rpc parameter is required")
		return
	}

	walletKeeper, err := wallet.NewWalletKeeper(walletCLI)
	if err != nil {
		fmt.Println(err)
		return
	}

	pubkey, err := walletKeeper.PublicKey()
	if err != nil {
		fmt.Println(err)
		return
	}

	client, err := client.NewClient(*rpcEndpoint)
	if err != nil {
		fmt.Println(err)
		return
	}

	balance, err := client.EthClient.BalanceAt(context.Background(), common.HexToAddress(pubkey), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Address balance:", balance)
}

// +++++++++++++++++++++++++++++++++++++++
// 		  BLOCK EXPLORER COMMANDS
// +++++++++++++++++++++++++++++++++++++++

func handleBlockCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	switch args[0] {
	case "number":
		executeBlockNumberCommand(args[1:])
	default:
		fmt.Println("Invalid command usage")
		return
	}
}

func executeBlockNumberCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	rpcEndpoint := findArg(args, "--rpc=")

	if rpcEndpoint == nil {
		fmt.Println("--rpc parameter is required")
		return
	}

	fmt.Println("Pulling the last block number from", rpcEndpoint, "...")
	explorer, err := apps.NewBlockExplorer(*rpcEndpoint)
	if err != nil {
		fmt.Println(err)
		return
	}

	number, err := explorer.BlockNumber()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(number)
}

// +++++++++++++++++++++++++++++++++++++++
// 		  	    APP COMMANDS
// +++++++++++++++++++++++++++++++++++++++

func handleAppCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	switch args[0] {
	case "soundminter":
		handlAppSoundminterCommand(args[1:])
	case "scoretable":
		handlAppScoretableCommand(args[1:])
	default:
		fmt.Println("Invalid command usage")
		return
	}
}

func handlAppSoundminterCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	switch args[0] {
	case "automint":
		executeAppSoundminterAutomint(args[1:])
	default:
		fmt.Println("Invalid command usage")
		return
	}
}

func executeAppSoundminterAutomint(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	rpcEndpoint := findArg(args, "--rpc=")

	if rpcEndpoint == nil {
		fmt.Println("--rpc parameter is required")
		return
	}

	soundminter, err := soundminter.NewSoundminter(
		*rpcEndpoint,
		walletCLI,
		common.HexToAddress("0xd19a5eE68e2ED7C19d509b6F4EcAd7409e79Ad58"),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = soundminter.Automint()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Execution completed")
}

func handlAppScoretableCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	switch args[0] {
	case "register-game":
		executeAppScoretableRegisterGameCommand(args[1:])
	default:
		fmt.Println("Invalid command usage")
		return
	}
}

func executeAppScoretableRegisterGameCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Invalid command usage")
		return
	}

	rpcEndpoint := findArg(args, "--rpc=")
	name := findArg(args, "--name=")
	meta := findArg(args, "--meta=")

	if rpcEndpoint == nil {
		fmt.Println("--rpc parameter is required")
		return
	}
	if name == nil {
		fmt.Println("--name parameter is required")
		return
	}
	if meta == nil {
		fmt.Println("--meta parameter is required")
		return
	}

	scoretable, err := scoretable.NewScoretable(*rpcEndpoint, walletCLI)
	if err != nil {
		fmt.Println(err)
		return
	}

	var metaBytes [32]byte
	copy(metaBytes[:], []byte(*meta))

	hash, err := scoretable.Registry.RegisterGame(*name, metaBytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Transaction mined: ", hash)
}

// +++++++++++++++++++++++++++++++++++++++
// 		  		   HELPERS
// +++++++++++++++++++++++++++++++++++++++

func findArg(args []string, target string) *string {
	for _, arg := range args {
		if strings.HasPrefix(arg, target) {
			result := strings.TrimPrefix(arg, target)
			return &result
		}
	}

	return nil
}
