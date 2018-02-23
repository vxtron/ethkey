package main

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

var maxAddresses int = 0
var numAddresses int = 0

func newKey() (string, string) {
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf(err.Error())
	}
	privateKey := crypto.FromECDSA(key)
	publicKey := crypto.FromECDSAPub(&key.PublicKey)
	address := crypto.Keccak256(publicKey[1:])

	return hex.EncodeToString(privateKey), hex.EncodeToString(address[len(address)-20:])
}

func generateKeys(c chan []string, patterns []string) {
	var matched bool

	for {
		privateKey, address := newKey()

		matched = false
		if len(patterns) == 0 {
			matched = true
		} else {
			for _, p := range patterns {
				if m, _ := path.Match(p, address); m {
					matched = true
					break
				}
			}
		}
		if matched {
			c <- []string{privateKey, address}
		}
	}
}

func handleOutput(c chan []string) {
	n := 0
	for v := range c {
		fmt.Printf("%s Key: %s Address: %s\n", time.Now().Format("[2006-01-02 15:04:05]"), v[0], v[1])
		n += 1
		if maxAddresses > 0 && n == maxAddresses {
			os.Exit(0)
		}
	}
}

func main() {
	cout := make(chan []string)

	app := cli.NewApp()
	app.Name = "ethkey"
	app.Usage = "Generate Ethereum addresses matching arbitrary patterns"
	app.Version = "1.0"

	app.ArgsUsage = "[WILDCARDS_TO_MATCH]..."
	app.Email = "vxtron@protonmail.com"
	app.HideHelp = true

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "num, n",
			Value: 5,
			Usage: "Number of addresses to create (0 = unlimited)",
		},
		cli.IntFlag{
			Name:  "threads, t",
			Value: runtime.GOMAXPROCS(0),
			Usage: "Number of threads to use",
		},
	}

	app.Action = func(c *cli.Context) error {
		t := c.Int("threads")
		if t <= 0 {
			t = runtime.GOMAXPROCS(0)
		}
		maxAddresses = c.Int("num")

		var patterns []string
		for _, p := range c.Args() {
			if _, err := path.Match(p, ""); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Not a valid wildcard pattern: %s\n", p)
				os.Exit(1)
			}
			patterns = append(patterns, strings.ToLower(p))
		}

		go handleOutput(cout)
		for i := 0; i < t; i = i + 1 {
			go generateKeys(cout, patterns)
		}
		select {} // Wait forever
		return nil
	}

	app.Run(os.Args)
}
