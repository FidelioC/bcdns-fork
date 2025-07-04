package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/khalidzahra/dns_client/eval"
	"github.com/khalidzahra/dns_client/substrate"
	"github.com/khalidzahra/dns_client/verifying_network"
)

func fetchSingleSpec(domain string, idx int, connector substrate.SubstrateInterface, eval, prefetch bool) (int, int64, *substrate.ChainSpecRes) {
	start := time.Now()
	target, err := connector.ResolveDomain(domain, eval)
	if err != nil {
		panic(err)
	}
	duration := time.Since(start).Milliseconds() // For evaluation
	if !prefetch {
		fmt.Println("================================================================")
		fmt.Println("		FOUND TARGET CHAIN SPEC")
		fmt.Println("================================================================")
		fmt.Printf("%+v\n", target)
	}
	return idx, duration, target
}

func fetchSpec(domain string, runs, runsPerSecond int, outFile string, evalFlag, useCache bool) (*substrate.ChainSpecRes) {
	resultChan := make(chan *eval.EvalResult)
	var resultArr []*eval.EvalResult
	var finalTarget *substrate.ChainSpecRes
	connector := substrate.NewSubstrateConnector(useCache)

	for i := 0; i < 32; i++ { // pre-fetch for caching
		_, _, _ = fetchSingleSpec(domain, i, connector, evalFlag, true)
	}

	var once sync.Once 

	eval.RunFuncPerSecond(func(currentRun int, wg *sync.WaitGroup) {
		idx, time, target := fetchSingleSpec(domain, currentRun, connector, evalFlag, false)
		once.Do(func() { // store only the first valid target
			finalTarget = target
		})
		wg.Done()
		resultChan <- &eval.EvalResult{Idx: idx, Time: time}
	}, runs, runsPerSecond)

	if evalFlag {
		for i := 0; i < runs; i++ {
			result := <-resultChan
			resultArr = append(resultArr, result)
		}

		eval.WriteToCSV(outFile, resultArr)
	}
	return finalTarget
}

func registerAssets(domain, outFile string, rps, totalRuns int) {
	connector := substrate.NewSubstrateConnector(true)
	resultsChan := make(chan string, 1000)
	nonce := connector.RegisterAsset(domain, fmt.Sprintf("asset%d", -1), 0, resultsChan)
	fmt.Printf("Initial nonce: %d\n", nonce)
	eval.RunFuncPerSecond(func(currentRun int, wg *sync.WaitGroup) {
		connector.RegisterAsset(domain, fmt.Sprintf("asset%d", currentRun), nonce+uint32(currentRun)+1, resultsChan)
		wg.Done()
	}, totalRuns, rps)

	var resultArr []*eval.EvalResult
	for i := 0; i < totalRuns; i++ {
		result := <-resultsChan
		resultParsed := strings.Split(result, ",")
		idx, _ := strconv.Atoi(resultParsed[0])
		time, _ := strconv.ParseInt(resultParsed[1], 10, 64)
		resultArr = append(resultArr, &eval.EvalResult{Idx: idx, Time: time})
	}

	eval.WriteToCSV(outFile, resultArr)
}

func listenToEvents(totalRuns int, outFile string, useCache bool) {
	connector := substrate.NewSubstrateConnector(useCache)
	resultsChan := make(chan string, 1000)
	go connector.ListenForEvents(resultsChan, true, 1000)

	var resultArr []*eval.EvalResult

	timeout := 30 * time.Second
listenerLoop:
	for i := 0; i < totalRuns; i++ {
		select {
		case result := <-resultsChan:
			resultParsed := strings.Split(result, ",")
			idx, _ := strconv.Atoi(resultParsed[0])
			timeVal, _ := strconv.ParseInt(resultParsed[1], 10, 64)
			resultArr = append(resultArr, &eval.EvalResult{Idx: idx, Time: timeVal})
		case <-time.After(timeout):
			fmt.Println("Timeout: No more events received for 30 seconds.")
			break listenerLoop
		}
	}

	eval.WriteToCSV(outFile, resultArr)
}

func main() {
	// Command line args
	var eval, assetEval, listen, useCache bool
	var runs, runsPerSecond int
	var domain, outFile string
	var target *substrate.ChainSpecRes
	var numNodes int = 3
	flag.StringVar(&domain, "domain", "example.com", "Domain to fetch chainspec for")
	flag.StringVar(&outFile, "outFile", "eval.csv", "Name of file to output eval results")
	flag.BoolVar(&eval, "eval", false, "Evaluate performance by running multiple times")
	flag.BoolVar(&assetEval, "assetEval", false, "Evaluate asset registration performance by running multiple times")
	flag.BoolVar(&listen, "listen", false, "Listen to events emitted by the chain")
	flag.BoolVar(&useCache, "useCache", false, "Use caching for interacting with the chain")
	flag.IntVar(&runs, "runs", 1, "Number of runs for evaluation")
	flag.IntVar(&runsPerSecond, "rps", 1, "Number of runs per second for evaluation")
	flag.Parse()

	if assetEval {
		registerAssets(domain, outFile, runsPerSecond, runs)
	} else if listen {
		listenToEvents(runs, outFile, useCache)
	} else {
		target = fetchSpec(domain, runs, runsPerSecond, outFile, eval, useCache)
	}

	// 1) create verifying network
	fmt.Println("TEST TARGET")
	fmt.Printf("\n%+v\n", target)
	json_target, err := verifying_network.ChainSpecToJson(target)
	if err != nil{
		panic(err)
	}
	fmt.Printf("json_target: %s", json_target)
	// 2) api call - client have verifying network nodes connect to the target, list of boot nodes
		// here, the verifying network will do logics to verify the boot nodes metadata
	nodes := verifying_network.CreateVerifierNetwork(numNodes)
	
	// TODO: 
	// check each i in the boot node array
	resultsChan := make(chan *verifying_network.VerificationResult, len(nodes))
	for i := range nodes {
		go func(i int) {
			// each of the result that's being returned by the node is a result of the consensus with the other peers
			res := nodes[i].GetBootNodeResult(context.Background(), json_target, 0, 30*time.Second)
			resultsChan <- res
		}(i)
	}

	var finalResult *verifying_network.VerificationResult
	for i := 0; i < numNodes; i++ {
		res := <-resultsChan
		fmt.Printf("Result %d:\n%+v\n", i, res)
		if res != nil && finalResult == nil {
			finalResult = res
		}
	}
	
	// TODO: 3) with the results, the client itself will do the boot node verification using zksnark?
	/* 
	json_target: {
		"id": "example",
		"bootNodes": [
			"/ip4/172.20.0.2/tcp/9945/p2p/12D3KooWNL4mZo8y7oAes3VRRnbHy91TDLxnjrDsnMFZkPebB2Rh",
			"/ip4/172.20.0.3/tcp/9945/p2p/12D3KooWNL4mZo8y7oAes3VRRnbHy91TDLxnjrDsnMFZkPebB2Rh"
		]
	}

	&{NodeID:12D3KooWF8u1fioBtHZGg9iciv893wbvmzCv4XBkjNwaUYhZazD1 ChainName:example NodeName:Substrate Node Version:0.1.0-7f7632fcc14 BlockHash:0x2af58600bbd7d3475ef1bd19ee413e04d3cfd4088b4e08434a27ae8132dfa135 BootIndex:0 IsSuccess:true ErrorMsg:}
	*/

	if finalResult == nil {
		fmt.Println("No valid verification result received.")
	} else {
		// Combine with already-parsed target
		combined := struct {
			ChainSpec          *substrate.ChainSpecRes                `json:"chainSpec"`
			VerificationResult verifying_network.VerificationResult `json:"verificationResult"`
		}{
			ChainSpec:          target,
			VerificationResult: *finalResult,
		}

		// Write to file
		file, err := os.Create("combined_result.json")
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(combined); err != nil {
			log.Fatalf("Failed to encode combined result: %v", err)
		}
		fmt.Println("Wrote combined_result.json successfully.")
	}

	select{} // prevent main from exiting
}
