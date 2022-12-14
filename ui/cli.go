package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// TODO: Propagate erorrs if needed

type ICLI interface {
	ReqChannel() chan UserInputRequest
	ResChannel() chan UserInputResponse
}

type CLI struct {
	reader     *bufio.Reader
	ReqChannel chan UserInputRequest
	ResChannel chan UserInputResponse
}

func NewCLI() *CLI {
	reader := bufio.NewReader(os.Stdin)
	reqChannel := make(chan UserInputRequest)
	resChannel := make(chan UserInputResponse)

	cli := &CLI{
		reader:     reader,
		ReqChannel: reqChannel,
		ResChannel: resChannel,
	}

	cli.bindIO()

	return cli
}

func (cli *CLI) bindIO() {
	go func() {
		request := <-cli.ReqChannel

		fmt.Printf("%s: ", request.title)
		defer fmt.Printf("\n")

		input, err := cli.reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}

		input = strings.Trim(input, "\r\n")

		cli.ResChannel <- UserInputResponse{input}
	}()
}
