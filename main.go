package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/0xNSHuman/dapp-tools/apps"
	"github.com/0xNSHuman/dapp-tools/apps/scoretable"
	"github.com/0xNSHuman/dapp-tools/ui"
	"github.com/0xNSHuman/dapp-tools/wallet"
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
	case "scoretable":
		handlAppScoretableCommand(args[1:])
	default:
		fmt.Println("Invalid command usage")
		return
	}
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
