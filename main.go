package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/tanishiking/btcwallet/key"
	"github.com/tanishiking/btcwallet/protocol"
)

func main() {
	usage := fmt.Sprintf(`
Usage of %s
	%s [SUBCOMMAND]
SUBCOMMAND
	show
		Show/Generate bitcoin address.
	balance
		Show balance.
	send <address> <amount> <fee>
		Send bitcoin.
`, os.Args[0], os.Args[0])

	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "balance":
		showBalance()
	case "show":
		generateNewBitcoinAddress()
	case "send":
		fmt.Println(len(os.Args))
		if len(os.Args) != 5 {
			fmt.Println(usage)
			os.Exit(1)
		}
		addr := os.Args[2]
		amount, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Printf("Invalid input amount %v\n", os.Args[3])
			fmt.Println(usage)
		}
		fee, err := strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Printf("Invalid input amount %v\n", os.Args[4])
			fmt.Println(usage)
		}
		sendBitcoin(addr, amount, fee)
	default:
		fmt.Println(usage)
	}
}

func showBalance() {
	protocol.Balance()
}

func sendBitcoin(addr string, amount int, fee int) {
	// protocol.Send("2N8hwP1WmJrFF5QWABn38y63uYLhnJYJYTF", 20000000, 10000000)
	protocol.Send(addr, amount, fee)
}

func generateNewBitcoinAddress() {
	privateKey, err := key.ReadOrGeneratePrivateKey()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	pubkey, err := key.GeneratePubKey(privateKey)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	btcAddr := key.EncodeBitcoinAddr(pubkey)
	fmt.Println(btcAddr)
}
