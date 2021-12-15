package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var genesisBlockUrl = "http://127.0.0.1:39149/"
var nodeData = make(map[string]string)

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func addMySelfToChain() {
	previousBlock := getPreviousBlock()
	requestObject := make(map[string]string)
	requestObject["ip"] = GetOutboundIP()
	requestJson, _ := json.Marshal(requestObject)
	http.Post(previousBlock+"/updateNext", "application/json", strings.NewReader(string(requestJson)))
	http.Get(previousBlock + "/next")
}

func getPreviousBlock() string {
	previousBlock := genesisBlockUrl
	for getNextOfBlock(genesisBlockUrl) != "end" {
		previousBlock = getNextOfBlock(genesisBlockUrl)
	}
	return previousBlock
}

func getNextOfBlock(blockUrl string) string {
	resp, err := http.Get(blockUrl + "/next")
	fmt.Printf("err: %v\n", err)
	resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("err: %v\n", err)
	return string(body)
}

func getSizeOfBlock(blockUrl string) string {
	resp, err := http.Get(blockUrl + "/size")
	fmt.Printf("err: %v\n", err)
	resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("err: %v\n", err)
	return string(body)
}

func searchForFreeNodes(genesisUrl string) string {
	freeNode := genesisUrl
	for getNextOfBlock(genesisUrl) != "end" {
		if getSizeOfBlock(genesisUrl) < "10" {
			freeNode = genesisUrl
		} else {
			genesisUrl = getNextOfBlock(genesisUrl)
		}
	}
	return freeNode
}

func addData(w http.ResponseWriter, req *http.Request) {
	var requestData map[string]string
	json.NewDecoder(req.Body).Decode(&requestData)
	if len(nodeData) < 10 {
		nodeData[requestData["id"]] = requestData["data"]
		return
	}
	freeNode := searchForFreeNodes(genesisBlockUrl)
	if genesisBlockUrl == "http://127.0.0.1:39149/" {
		fmt.Fprint(w, "Cannot Add Values,No more nodes in Network")
		return
	}
	// fmt.Fprint(w, freeNode)
	fmt.Print(freeNode)
	//TODO: make a post request to freeNode
}

func updateNext(w http.ResponseWriter, req *http.Request) {
	var requestData map[string]string
	json.NewDecoder(req.Body).Decode(&requestData)
	nodeData["next"] = requestData["ip"]
	fmt.Fprint(w, "Node Successfully added to Chain")
}

func sayFileLimit(w http.ResponseWriter, req *http.Request) {
	size := len(nodeData)
	fmt.Fprint(w, size)
}

func sayNextBlock(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, nodeData["next"])
}

// func getData(w http.ResponseWriter, req *http.Request) {

// }

func GetRandomId(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func main() {
	if len(os.Args) >= 2 {
		genesisBlockUrl = os.Args[1:][0]
		addMySelfToChain()
	}

	// generate needed data for the node
	nodeId := GetRandomId(10)

	// assign default nodevalues
	nodeData["id"] = nodeId
	nodeData["next"] = "end"

	// function handling
	http.HandleFunc("/size", sayFileLimit)
	http.HandleFunc("/add", addData)
	// http.HandleFunc("/get", headers)
	http.HandleFunc("/next", sayNextBlock)
	http.HandleFunc("/updateNext", updateNext)

	log.Fatal(http.ListenAndServe(":39149", nil))
}
