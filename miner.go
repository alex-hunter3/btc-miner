package main

import (
	"crypto/sha256"
	"io/ioutil"
	"encoding/hex"
	"context"
	"strconv"
	"time"
	"math"
	"fmt"
)

type Block struct {
	blockNumber  int
	transactions string
	previousHash string
	newHash      string
	nonce        uint64
	difficulty   int
	timeTaken    time.Duration
}

func encrypt(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func startswith(text string, prefix string) bool {
	for i := 0; i < len(prefix); i++ {
		if text[i] != prefix[i] {
			return false
		}
	}

	return true
}

func miner(ctx context.Context, blockNumber int, transactions string, previousHash string, zeroPrefix string, startNonce uint64, nonceChan chan uint64, hashChan chan string) {
	var text string
	var newHash string

	for {
		select {

		case <- ctx.Done():
			return
		default:
			text = strconv.Itoa(blockNumber) + transactions + previousHash + strconv.FormatUint(startNonce, 10)
			newHash = encrypt(text)

			if startswith(newHash, zeroPrefix) {
				nonceChan <- startNonce
				hashChan  <- newHash

				close(nonceChan)
				close(hashChan)
				break
			} else {
				startNonce++
			}
		}
	}
}

func mine(blockNumber int, transactions string, previousHash string, zeroPrefix int) Block {
	var prefixString string
	var newHash string
	var nonce uint64
	var startNonce uint64
	var numMiners float64 = 6

	nonceChan := make(chan uint64)
	hashChan := make(chan string)

	for i := 0; i < zeroPrefix; i++ {
		prefixString += "0"
	}

	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < int(numMiners); i++{
		startNonce = uint64((float64(i) / numMiners) * math.Pow(2, 64))
		fmt.Println("Started miner with start nonce of", startNonce)
		go miner(ctx, blockNumber, transactions, previousHash, prefixString, startNonce, nonceChan, hashChan)
	}

	nonce = <- nonceChan
	newHash = <- hashChan
	cancel()

	block := Block{
		blockNumber,
		transactions,
		previousHash,
		newHash,
		nonce,
		zeroPrefix,
		time.Since(start),
	}

	return block
}

func write_file(block Block) {
	var text string

	text = strconv.Itoa(block.blockNumber) + "," + block.transactions + "," + block.previousHash + "," + block.newHash + "," + strconv.FormatUint(block.nonce, 10) + "," + strconv.Itoa(block.difficulty) + "," + block.timeTaken.String()

	err := ioutil.WriteFile("miner_log.txt", []byte(text + "\n"), 0644)

	if err != nil {
		panic(err)
	}
}

func main() {
	var ledger string = "Alex->Emma->20"
	var previousHash string = "00000000000000000007c5d3134fb591090eef0e123a26ca6148ba0c59b0a428"
	var difficulty int = 19
	var blockNumber int = 2

	// All the data is dummy data to be tested with
	// The current zeros difficulty of bitcoin as of 22/01/2021 is 19 zeros

	fmt.Println("Mining has begun")
	block := mine(blockNumber, ledger, previousHash, difficulty)
	fmt.Println("Finished mining")
	fmt.Println("New block hash: " + block.newHash)
	fmt.Println("Nonce: " + strconv.FormatUint(block.nonce, 10))
	fmt.Println("Time taken:", block.timeTaken)

	write_file(block)
}
