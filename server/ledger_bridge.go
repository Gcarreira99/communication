package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID        = "Org1MSP"
	cryptoPath   = "../../../go/src/github.com/47892704/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com"
	certPath     = cryptoPath + "/users/User1@org1.example.com/msp/signcerts/cert.pem"
	keyPath      = cryptoPath + "/users/User1@org1.example.com/msp/keystore/"
	tlsCertPath  = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint = "localhost:7051"
	gatewayPeer  = "peer0.org1.example.com"
)

var contract *client.Contract
var clientConnection *grpc.ClientConn
var gw *client.Gateway

type FabricResult struct {
	Key   string `json:"Key"`
	Asset string `json:"Asset"`
}

var startKey = 1000000
var endKey = 1000000

func blockchainConnectionStartup() {
	// The gRPC client connection should be shared by all Gateway connections to this endpoint
	clientConnection = newGrpcConnection()

	id := newIdentity()
	sign := newSign()
	// Create a Gateway connection for a specific client identity
	var err error
	gw, err = client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "basic"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)
	contract = network.GetContract(chaincodeName)
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)
	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	//Prints current directory
	dir, err := os.Getwd()
	log.Print("-> Ledger Bridge current directory: ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(dir)

	certificate, err := loadCertificate(certPath)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	files, err := os.ReadDir(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key directory: %w", err))
	}
	privateKeyPEM, err := os.ReadFile(path.Join(keyPath, files[0].Name()))

	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

// Evaluate a transaction to query ledger state.
func getAllAssets() []FabricResult {
	fmt.Println("\n--> Evaluate Transaction: GetAllAssets, function returns all the current assets on the ledger")
	var results []FabricResult
	evaluateResult, err := contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	fmt.Println(evaluateResult)
	if len(evaluateResult) < 1 {
		return nil
	}
	err2 := json.Unmarshal(evaluateResult, &results)
	if err2 != nil {
		panic(fmt.Errorf("failed to unmarshal: %w", err))
	}

	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result:%s\n", result)
	return results
}

func createAsset(query string) {
	//fmt.Printf("\n--> Submit Transaction: CreateAsset, create new state with ID and query arguments \n")
	_, err := contract.SubmitTransaction("CreateAsset", strconv.Itoa(endKey), query)

	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}
	fmt.Printf("*** Transaction committed successfully\n")
	endKey += 1
}

func getByRangeAsset() []FabricResult {
	fmt.Printf("\n--> Evaluate Transaction: GetByRangeAsset\n")
	var results []FabricResult
	if startKey == endKey {
		fmt.Printf("StartKey & EndKey: SAME VALUE!")
		return nil
	}
	fmt.Printf("startKey: %s; endKey: %s\n", strconv.Itoa(startKey), strconv.Itoa(endKey))
	evaluateResult, err := contract.EvaluateTransaction("GetByRange", strconv.Itoa(startKey), strconv.Itoa(endKey))
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	if len(evaluateResult) < 1 {
		fmt.Println("EMPTY BLOCKCHAIN!")
		return nil
	}
	err2 := json.Unmarshal(evaluateResult, &results)
	if err2 != nil {
		panic(fmt.Errorf("failed to unmarshal: %w", err))
	}
	result := formatJSON(evaluateResult)
	startKey = endKey
	fmt.Printf("*** Result:%s\n", result)
	return results
}

// Format JSON data
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}
