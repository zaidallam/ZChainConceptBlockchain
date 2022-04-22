package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Block struct {
	Nonce     uint
	PrevHash  string
	Data      []byte
	Hash      string
	Timestamp string
}

type ValidationRequest struct {
	Block           Block
	DiscoveredNodes []string
}

type Vote struct {
	Vote            bool
	DiscoveredNodes []string
}

type newNode struct {
	Node string
}

type Message struct {
	Data string
}

var ZChain []Block
var DiscoveredNodeAddresses []string

func assembleBlock(prevBlock Block, data []byte) Block {
	var newBlock Block

	newBlock.Nonce = prevBlock.Nonce + 1
	newBlock.PrevHash = prevBlock.Hash
	newBlock.Data = data
	newBlock.Timestamp = time.Now().String()

	newBlock.Hash = hashBlock(newBlock)

	return newBlock
}

func hashBlock(block Block) string {
	blockBytes := []byte(fmt.Sprint(block.Nonce, block.Timestamp, block.Data, block.PrevHash))
	sha256 := sha256.New()
	sha256.Write(blockBytes)

	return hex.EncodeToString(sha256.Sum(nil))
}

func isValidBlock(block Block, prevBlock Block) bool {
	if prevBlock.Nonce+1 != block.Nonce ||
		prevBlock.Hash != block.PrevHash ||
		hashBlock(block) != block.Hash {
		return false
	}

	return true
}

func overwriteChain(newChain []Block) {
	ZChain = newChain
}

func createRouter() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", getBlockchain).Methods("GET")
	router.HandleFunc("/", postBlock).Methods("POST")
	router.HandleFunc("/exists", exists).Methods("GET")
	router.HandleFunc("/nodes", getDiscoveredNodes).Methods("GET")
	router.HandleFunc("/nodes", postDiscoveredNode).Methods("POST")
	router.HandleFunc("/validate", validateBlock).Methods("POST")

	return router
}

func exists(responseWriter http.ResponseWriter, request *http.Request) {
	respond(responseWriter, request, 200, true)
}

func validateBlock(responseWriter http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var message ValidationRequest
	json.NewDecoder(request.Body).Decode(&message)
	isValid := isValidBlock(message.Block, ZChain[len(ZChain)-1])

	if isValid {
		newBlockchain := append(ZChain, message.Block)
		overwriteChain(newBlockchain)
	}

	for _, node := range message.DiscoveredNodes {
		tryAddNode(node)
	}
	respond(responseWriter, request, 200, Vote{isValid, DiscoveredNodeAddresses})
}

func getBlockchain(responseWriter http.ResponseWriter, request *http.Request) {
	bytes, err := json.Marshal(ZChain)
	if err == nil {
		io.WriteString(responseWriter, string(bytes))
		return
	}

	http.Error(responseWriter, err.Error(), 500)
}

func getDiscoveredNodes(responseWriter http.ResponseWriter, request *http.Request) {
	bytes, err := json.Marshal(DiscoveredNodeAddresses)
	if err == nil {
		io.WriteString(responseWriter, string(bytes))
		return
	}

	http.Error(responseWriter, err.Error(), 500)
}

func postDiscoveredNode(responseWriter http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var message newNode
	json.NewDecoder(request.Body).Decode(&message)

	if validateNode(message.Node) {
		tryAddNode(message.Node)
	}
}

func validateNode(node string) bool {
	response, err := http.Get(node + "/exists")

	if err != nil {
		return false
	}

	var res bool

	json.NewDecoder(response.Body).Decode(&res)

	return res
}

func postBlock(responseWriter http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	if len(DiscoveredNodeAddresses) < 1 {
		respond(responseWriter, request, 500, "ERROR: Block not created because no peers have been discovered; retrieving consensus is impossible.")
	}

	var message Message
	json.NewDecoder(request.Body).Decode(&message)

	newestBlock := ZChain[len(ZChain)-1]
	newBlock := assembleBlock(newestBlock, []byte(message.Data))

	if isValidBlock(newBlock, newestBlock) {
		if hasConsensus(newBlock) {
			newBlockchain := append(ZChain, newBlock)
			overwriteChain(newBlockchain)
			respond(responseWriter, request, 201, newBlock)
			return
		}
	}

	respond(responseWriter, request, 400, "ERROR: Block not created")
}

func hasConsensus(newBlock Block) bool {
	currentCallNewNodes := DiscoveredNodeAddresses
	httpClient := http.Client{
		Timeout: 30 * time.Second,
	}
	var validations float32
	var totalVotes float32

	for len(currentCallNewNodes) > 0 {
		var wg sync.WaitGroup

		votes := make(chan bool, len(currentCallNewNodes))
		newNodes := make(chan []string, len(currentCallNewNodes))

		for _, s := range currentCallNewNodes {
			wg.Add(1)
			go retrieveConsensusVote(&wg, s, newBlock, &httpClient, votes, newNodes)
		}

		wg.Wait()
		close(votes)
		close(newNodes)

		currentCallNewNodes = []string{}
		for nodes := range newNodes {
			for _, node := range nodes {
				if tryAddNode(node) {
					currentCallNewNodes = append(currentCallNewNodes, node)
				}
			}
		}

		for val := range votes {
			if val {
				validations++
			}

			totalVotes++
		}
	}

	return validations/totalVotes > 0.8
}

func retrieveConsensusVote(wg *sync.WaitGroup, nodeAddress string, newBlock Block, httpClient *http.Client, votes chan bool, newNodes chan []string) {
	defer wg.Done()

	request := ValidationRequest{newBlock, DiscoveredNodeAddresses}
	data, err := json.Marshal(request)

	if err != nil {
		votes <- false
		return
	}

	response, err := httpClient.Post(nodeAddress+"/validate", "application/json", bytes.NewBuffer(data))

	if err != nil {
		votes <- false
		return
	}

	var res Vote

	json.NewDecoder(response.Body).Decode(&res)

	votes <- res.Vote
	newNodes <- res.DiscoveredNodes
}

func tryAddNode(node string) bool {
	thisNode := "http://localhost:" + os.Getenv("PORT")

	if node == thisNode {
		return false
	}

	exists := false

	for _, discNode := range DiscoveredNodeAddresses {
		if discNode == node {
			exists = true
		}
	}

	if exists {
		return false
	}

	DiscoveredNodeAddresses = append(DiscoveredNodeAddresses, node)
	return true
}

func respond(responseWriter http.ResponseWriter, request *http.Request, statusCode int, body interface{}) {
	response, err := json.Marshal(body)
	if err != nil {
		responseWriter.WriteHeader(500)
		responseWriter.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}

	responseWriter.WriteHeader(statusCode)
	responseWriter.Write(response)
}

func startNode(port string, errChan chan error) {
	spew.Print("Starting ZChain...")
	router := createRouter()
	server := &http.Server{
		Handler:        router,
		MaxHeaderBytes: 1 << 20,
		Addr:           ":" + port,
		WriteTimeout:   10 * time.Second,
		ReadTimeout:    10 * time.Second,
	}

	spew.Print("Listening on port " + port)
	errChan <- server.ListenAndServe()
}

func main() {
	godotenv.Load()

	go func() {
		genesisBlock := Block{0, "", nil, "", time.Now().String()}
		ZChain = append(ZChain, genesisBlock)
	}()

	discoveryNode := os.Getenv("DISCOVERY_NODE_ADDRESS")
	port := os.Getenv("PORT")
	errChan := make(chan error)

	go startNode(port, errChan)

	if discoveryNode != "" {
		response, err := http.Get(discoveryNode + "/")
		tryAddNode(discoveryNode)

		if err != nil {
			return
		}

		var registerSelfNode = newNode{"http://localhost:" + port}
		data, err := json.Marshal(registerSelfNode)

		if err != nil {
			return
		}

		http.Post(discoveryNode+"/nodes", "application/json", bytes.NewBuffer(data))

		var res []Block

		json.NewDecoder(response.Body).Decode(&res)

		overwriteChain(res)
	}

	spew.Print(<-errChan)
}
