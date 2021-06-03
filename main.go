package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/HTaeha/Blockchain-in-Golang/blockchain"
)

// CommandLine : CommandLine으로 원하는 동작을 실행시킬 수 있도록 한다.
type CommandLine struct {
	blockchain *blockchain.BlockChain
}

// printUsage : print cli의 사용법을 알려준다.
func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" add -block BLOCK_DATA - add a block to the chain")
	fmt.Println(" print - Prints the blocks in the chain")
}

// validateArgs : Argument를 검증한다.
func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		// runtime.Goexit은 현재 goroutine을 종료시킨다.
		// main 프로그램은 정상적으로 돌기 때문에 DB가 충돌을 일으키지 않고 종료할 수 있다.
		// os.exit()을 사용하면 프로그램 자체를 종료시키기 때문에 DB가 정상종료되지 않을 수 있다.
		runtime.Goexit()
	}
}

// addBlock : cli를 통해 블록을 추가한다.
func (cli *CommandLine) addBlock(data string) {
	cli.blockchain.AddBlock(data)
	fmt.Println("Added Block!")
}

// printChain : 블록체인에서 마지막부터 첫번째 블록 내용을 출력한다.
func (cli *CommandLine) printChain() {
	iter := cli.blockchain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		// Validate 과정은 매우 빠르게 처리된다.
		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

// run Command Line Interface.
func (cli *CommandLine) run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")

	switch os.Args[1] {
	// add command 일 때 파싱
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	// print command 일 때 파싱
	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if addBlockCmd.Parsed() {
		// addBlockData가 없으면 Usage() 실행 후 종료.
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		// 블록 추가.
		cli.addBlock(*addBlockData)
	}

	// print command 일 때 printChain()을 실행.
	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func main() {
	defer os.Exit(0)
	chain := blockchain.InitBlockChain()
	// 메인이 종료되기 전에 DB를 종료.
	defer chain.Database.Close()

	cli := CommandLine{chain}
	cli.run()
}
