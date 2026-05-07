package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/NichiNect/cachedist/sdk"
)

func main() {
	setCmd := flag.NewFlagSet("set", flag.ExitOnError)
	setNodes := setCmd.String("nodes", "localhost:7001", "comma-separated list of node addresses")
	setTTL := setCmd.Int("ttl", 60, "TTL for the key in seconds")

	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getNodes := getCmd.String("nodes", "localhost:7001", "comma-separated list of node addresses")

	if len(os.Args) < 2 {
		fmt.Println("expected 'set' or 'get' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "set":
		setCmd.Parse(os.Args[2:])
		if setCmd.NArg() < 2 {
			fmt.Println("usage: cli set --nodes=... --ttl=... <key> <value>")
			os.Exit(1)
		}
		key := setCmd.Arg(0)
		value := setCmd.Arg(1)
		nodes := strings.Split(*setNodes, ",")

		client := sdk.NewClient(nodes)
		err := client.Set(key, value, *setTTL)
		if err != nil {
			fmt.Printf("Error setting key: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully set key '%s' to '%s'\n", key, value)

	case "get":
		getCmd.Parse(os.Args[2:])
		if getCmd.NArg() < 1 {
			fmt.Println("usage: cli get --nodes=... <key>")
			os.Exit(1)
		}
		key := getCmd.Arg(0)
		nodes := strings.Split(*getNodes, ",")

		client := sdk.NewClient(nodes)
		val, found, err := client.Get(key)
		if err != nil {
			fmt.Printf("Error getting key: %v\n", err)
			os.Exit(1)
		}
		if !found {
			fmt.Println("Key not found")
		} else {
			fmt.Printf("Value: %s\n", val)
		}

	default:
		fmt.Println("expected 'set' or 'get' subcommands")
		os.Exit(1)
	}
}
