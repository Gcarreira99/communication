package main

import (
	"communication/service"
	context2 "context"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var ctx context.Context
var driver neo4j.DriverWithContext

var port = flag.Int("port", 3333, "the port to serve on")

// should implement the interface myPkgName.ServiceServer
type serviceServer struct {
	// type embedded to comply with Google lib
	service.UnimplementedServiceServer
}

type Component struct {
	Category string `json:"category"`
	Hash     string `json:"hash"`
}

type Asset struct {
	Nodes         []Component `json:"nodes"`
	Relationships []Component `json:"relationships"`
}

func hashConvert(input []byte) string {
	hasher := sha1.New()
	hasher.Write(input)
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func checkNodes(command string) []string {
	//check nodes present in the command
	nodeList := []string{"LINK", "SOURCE", "TWEET", "HASHTAG", "USER"}
	var itNotContains bool
	for i := 0; i < len(nodeList); i++ {
		itNotContains = strings.Contains(strings.ToUpper(command), nodeList[i])
		if itNotContains {
			//fmt.Println("Does not contain node: %v", nodeList[i])
			nodeList = append(nodeList[:i], nodeList[i+1:]...)
		}
	}
	return nodeList
}

func checkRelationships(command string) []string {
	//check relationships present in the command
	relationshipsList := []string{"CONTAINS", "TAGS", "POSTS", "FOLLOWS", "MENTIONS", "RETWEETS", "USING"}
	var itNotContains bool
	for i := 0; i < len(relationshipsList); i++ {
		itNotContains = strings.Contains(strings.ToUpper(command), relationshipsList[i])
		if itNotContains {
			//fmt.Println("Does not contain relationship: %v", relationshipsList[i])
			relationshipsList = append(relationshipsList[:i], relationshipsList[i+1:]...)
		}
	}
	return relationshipsList
}

func checkDatabaseStartup() (data Asset) {
	nodes := []string{"Link", "Source", "Tweet", "Hashtag", "User"}
	relationships := []string{"CONTAINS", "TAGS", "POSTS", "FOLLOWS", "MENTIONS", "RETWEETS", "USING"}
	var nodeArray []Component
	var relationshipArray []Component
	//Check blockchain if the nodes hashes are exactly the same, checks by node category
	for i := 0; i < len(nodes); i++ {
		query := "MATCH (n: " + nodes[i] + ") RETURN n"
		result, err := neo4j.ExecuteQuery(ctx, driver, query,
			nil,
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase("neo4j"))
		if err != nil {
			panic(err)
		}

		resultJson, _ := json.Marshal(result.Records)
		resultHash := hashConvert(resultJson)
		nodeState := Component{Category: nodes[i], Hash: resultHash}
		nodeArray = append(nodeArray, nodeState)
		//nodeHashes = append(nodeHashes, resultHash)
	}
	//Check blockchain if the relationships hashes are exactly the same
	for i := 0; i < len(relationships); i++ {
		query := "MATCH p=()-[r:" + relationships[i] + "]->() RETURN p"
		result, err := neo4j.ExecuteQuery(ctx, driver, query,
			nil,
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase("neo4j"))
		if err != nil {
			panic(err)
		}

		resultJson, _ := json.Marshal(result.Records)
		resultHash := hashConvert(resultJson)
		relationshipState := Component{Category: relationships[i], Hash: resultHash}
		relationshipArray = append(relationshipArray, relationshipState)
		//relationHashes = append(relationHashes, resultHash)
	}
	assetData := Asset{
		Nodes:         nodeArray,
		Relationships: relationshipArray,
	}
	return assetData
}

func checkDatabase(command string) (data Asset) {
	var nodes []string
	var relationships []string
	nodes = checkNodes(command)
	relationships = checkRelationships(command)
	fmt.Println("Command nodes are: %v", nodes)
	fmt.Println("Command relationships are: %v", relationships)
	var nodeArray []Component
	var relationshipArray []Component
	//Reading database to check blockchain if the nodes hashes are exactly the same, checks by node category
	for i := 0; i < len(nodes); i++ {
		result, err := neo4j.ExecuteQuery(ctx, driver,
			command,
			nil,
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase("neo4j"))
		if err != nil {
			panic(err)
		}

		resultJson, _ := json.Marshal(result.Records)
		resultHash := hashConvert(resultJson)
		nodeState := Component{Category: nodes[i], Hash: resultHash}
		nodeArray = append(nodeArray, nodeState)

	}
	///Reading database to check blockchain if the relationships hashes are exactly the same
	for i := 0; i < len(relationships); i++ {
		result, err := neo4j.ExecuteQuery(ctx, driver,
			command,
			nil,
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase("neo4j"))
		if err != nil {
			panic(err)
		}

		resultJson, _ := json.Marshal(result.Records)
		resultHash := hashConvert(resultJson)
		relationshipState := Component{Category: relationships[i], Hash: resultHash}
		relationshipArray = append(relationshipArray, relationshipState)
	}

	assetData := Asset{
		Nodes:         nodeArray,
		Relationships: relationshipArray,
	}

	return assetData
}

func validateDatabaseBlockchain(componentsDatabase []Component, componentsBlockchain []Component) bool {
	for i := 0; i < len(componentsDatabase); i++ {
		category := componentsDatabase[i].Category
		hash := componentsDatabase[i].Hash
		if (category != "") && (hash != "") {
			if (category != componentsBlockchain[i].Category) || (hash != componentsBlockchain[i].Hash) {
				fmt.Printf("Database Component category: %s\n", category)
				fmt.Printf("Blockchain Component category: %s\n", componentsBlockchain[i].Category)
				fmt.Printf("Database Component hash: %s\n", hash)
				fmt.Printf("Blockchain Component hash: %s\n", componentsBlockchain[i].Hash)
				return false
			}
		}
	}
	return true
}

func checkSystemPortion(command string) (bool, Asset) {
	dataDatabase := checkDatabase(command)
	dataBlockchain := readLastAssetByID()

	if (validateDatabaseBlockchain(dataDatabase.Nodes, dataBlockchain.Nodes) == true) && (validateDatabaseBlockchain(dataDatabase.Relationships, dataBlockchain.Relationships) == true) {
		return true, dataDatabase
	}
	return false, Asset{}
}

func checkSystemWhole() (bool, Asset) {
	dataDatabase := checkDatabaseStartup()
	dataBlockchain := readLastAssetByID()

	if (validateDatabaseBlockchain(dataDatabase.Nodes, dataBlockchain.Nodes) == true) && (validateDatabaseBlockchain(dataDatabase.Relationships, dataBlockchain.Relationships) == true) {
		return true, dataDatabase
	}
	return false, Asset{}
}

func connectDatabase() (context.Context, neo4j.DriverWithContext) {
	ctx := context.Background()
	// URI examples: "neo4j://localhost", "neo4j+s://xxx.databases.neo4j.io"
	dbUri := "neo4j://localhost:7687"
	dbUser := "neo4j"
	dbPassword := "password"
	driver, err := neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""))
	if err != nil {
		panic(err)
	}
	//defer driver.Close(ctx)

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connection established.")

	//session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "testneo4j"})
	//defer session.Close(ctx)

	return ctx, driver
}

func disconnectDatabase() {
	defer func(driver neo4j.DriverWithContext, ctx context2.Context) {
		err := driver.Close(ctx)
		if err != nil {

		}
	}(driver, ctx)
}

// WriteDatabase Create a node representing a person named Alice
func (m *serviceServer) WriteDatabase(ctx_c context.Context, request *service.WriteDatabaseRequest) (*service.WriteDatabaseResponse, error) {
	valid, data := checkSystemPortion(request.Value)
	if valid == false {
		return &service.WriteDatabaseResponse{Value: "WRITE NOT SUCCESS"}, nil
	}

	createAsset(data)
	result, err := neo4j.ExecuteQuery(ctx, driver,
		request.Value,
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created %v nodes in %+v.\n",
		result.Summary.Counters().NodesCreated(),
		result.Summary.ResultAvailableAfter())
	return &service.WriteDatabaseResponse{Value: "WRITE SUCCESS"}, nil
}

// ReadDatabase Retrieve all Person nodes
func (m *serviceServer) ReadDatabase(ctx_c context.Context, request *service.ReadDatabaseRequest) (*service.ReadDatabaseResponse, error) {
	if request.Value == "all" {
		getAllAssets()
		return &service.ReadDatabaseResponse{Value: "READ SUCCESS"}, nil
	} else if request.Value == "last" {
		fmt.Println("[ReadDatabase/last]")
		data, _ := checkSystemWhole()
		dataJson, _ := json.Marshal(data)
		return &service.ReadDatabaseResponse{Value: "[READ SUCCESS]\n" + string(dataJson)}, nil
	}
	_ = checkDatabase(request.Value)
	result, err := neo4j.ExecuteQuery(ctx, driver,
		request.Value,
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		panic(err)
	}

	// Loop through results and do something with them
	fmt.Println(result.Summary)
	for _, record := range result.Records {
		name, _ := record.Get("h.name") // .Get() 2nd return is whether key is present
		fmt.Println(name)
		// or
		// fmt.Println(record.AsMap())  // get Record as a map
	}

	// Summary information
	fmt.Printf("The query `%v` returned %v records in %+v.\n",
		result.Summary.Query().Text(), len(result.Records),
		result.Summary.ResultAvailableAfter())

	//Prints current directory
	path, err := os.Getwd()
	log.Print("-> Server current directory: ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)

	return &service.ReadDatabaseResponse{Value: "READ SUCCESS"}, nil
}

// UpdateDatabase Update node Alice to add an age property
func (m *serviceServer) UpdateDatabase(ctx_c context.Context, request *service.UpdateDatabaseRequest) (*service.UpdateDatabaseResponse, error) {
	result, err := neo4j.ExecuteQuery(ctx, driver, `
    MATCH (p:Person {name: $name})
    SET p.age = $age
    `, map[string]any{
		"name": "Alice",
		"age":  42,
	}, neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		panic(err)
	}
	fmt.Println("Query updated the database?",
		result.Summary.Counters().ContainsUpdates())

	return &service.UpdateDatabaseResponse{Value: "UPDATE SUCCESS"}, nil
}

// DeleteDatabase Remove the Alice node
func (m *serviceServer) DeleteDatabase(ctx_c context.Context, request *service.DeleteDatabaseRequest) (*service.DeleteDatabaseResponse, error) {
	// This does not delete _only_ p, but also all its relationships!
	result, err := neo4j.ExecuteQuery(ctx, driver, `
    MATCH (p:Person {name: $name})
    DETACH DELETE p
    `, map[string]any{
		"name": "Alice",
	}, neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		panic(err)
	}
	fmt.Println("Query updated the database?",
		result.Summary.Counters().ContainsUpdates())

	return &service.DeleteDatabaseResponse{Value: "DELETE SUCCESS"}, nil
}

func (m *serviceServer) Create(ctx_c context.Context, request *service.CreateRequest) (*service.CreateResponse, error) {
	log.Println("Create called")
	return &service.CreateResponse{Pdf: []byte("RESPONSE")}, nil
}

func loadCertificates() credentials.TransportCredentials {
	// read ca's cert, verify to client's certificate
	caPem, err := ioutil.ReadFile("/keys/cert/ca-cert.pem")
	if err != nil {
		log.Fatal(err)
	}

	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		log.Fatal(err)
	}

	// read server cert & key
	serverCert, err := tls.LoadX509KeyPair("/keys/cert/server-cert.pem", "/keys/cert/server-key.pem")
	if err != nil {
		log.Fatal(err)
	}

	// configuration of the certificate what we want to
	conf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	//create tls certificate
	tlsCredentials := credentials.NewTLS(conf)

	return tlsCredentials
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create tls based credential.
	creds := loadCertificates()

	s := grpc.NewServer(grpc.Creds(creds))
	myServiceServer := &serviceServer{}
	service.RegisterServiceServer(s, myServiceServer)
	log.Printf("server listening at %v", lis.Addr())
	//Creates a database connection
	ctx, driver = connectDatabase()

	//Create a blockchain connection
	blockchainConnectionStartup()
	//Add first state in the blockchain with the startup data
	assetData := checkDatabaseStartup()
	createAsset(assetData)

	//Handles CTRL^C signal to execute a graceful exit by closing the database connection
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		s := <-c
		if s == syscall.SIGINT {
			fmt.Print(" ")
			fmt.Println("Detected")
			fmt.Println("Closing database & blockchain connection...")
			disconnectDatabase()
			err := gw.Close()
			if err != nil {
				return
			}
			err = clientConnection.Close()
			if err != nil {
				return
			}
			fmt.Println("Connections closed with success")
			os.Exit(0)
		}
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
