package ui

type WalletCLI struct {
	cli *CLI
}

func NewWalletCLI() *WalletCLI {
	cli := NewCLI()

	return &WalletCLI{cli: cli}
}

func (wcli *WalletCLI) EnterPassphrase() (string, error) {
	wcli.cli.ReqChannel <- UserInputRequest{"Enter wallet passphrase"}
	res := <-wcli.cli.ResChannel

	return res.text, nil
}
