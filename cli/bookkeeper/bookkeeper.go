package bookkeeper

import (
	. "BOT/cli/common"
	"BOT/core/transaction"
	"BOT/crypto"
	"BOT/net/httpjsonrpc"
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/urfave/cli"
)

func makeBookkeeperTransaction(pubkey *crypto.PubKey, op bool, cert []byte) (string, error) {
	tx, _ := transaction.NewBookKeeperTransaction(pubkey, op, cert)
	attr := transaction.NewTxAttribute(transaction.Nonce, []byte(strconv.FormatInt(rand.Int63(), 10)))
	tx.Attributes = make([]*transaction.TxAttribute, 0)
	tx.Attributes = append(tx.Attributes, &attr)
	var buffer bytes.Buffer
	if err := tx.Serialize(&buffer); err != nil {
		fmt.Println("serialize bookkeeper transaction failed")
		return "", err
	}
	return hex.EncodeToString(buffer.Bytes()), nil
}

func assetAction(c *cli.Context) error {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	var pubkeyHex []byte
	var err error
	var add bool
	addPubkey := c.String("add")
	subPubkey := c.String("sub")
	if addPubkey == "" && subPubkey == "" {
		fmt.Println("missing --add or --sub")
		return nil
	}

	if addPubkey != "" {
		pubkeyHex, err = hex.DecodeString(addPubkey)
		add = true
	}
	if subPubkey != "" {
		if pubkeyHex != nil {
			fmt.Println("using --add or --sub")
			return nil
		}
		pubkeyHex, err = hex.DecodeString(subPubkey)
		add = false
	}
	if err != nil {
		fmt.Println("Invalid public key in hex")
		return nil
	}
	pubkey, err := crypto.DecodePoint(pubkeyHex)
	if err != nil {
		fmt.Println("Invalid public key")
		return nil
	}
	cert := c.String("cert")
	txHex, err := makeBookkeeperTransaction(pubkey, add, []byte(cert))
	if err != nil {
		return err
	}
	resp, err := httpjsonrpc.Call(Address(), "sendrawtransaction", 0, []interface{}{txHex})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	FormatOutput(resp)
	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "bookkeeper",
		Usage:       "add or remove bookkeeper",
		Description: "With nodectl bookkeeper, you could add or remove bookkeeper.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "add, a",
				Usage: "add a bookkeeper",
			},
			cli.StringFlag{
				Name:  "sub, s",
				Usage: "sub a bookkeeper",
			},
			cli.StringFlag{
				Name:  "cert, c",
				Usage: "authorized certificate",
			},
		},
		Action: assetAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			PrintError(c, err, "bookkeeper")
			return cli.NewExitError("", 1)
		},
	}
}