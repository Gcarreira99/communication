package main

import (
	"communication/service"
	context2 "context"
	"crypto/tls"
	"crypto/x509"
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
	"syscall"
	"time"
)

var ctx context.Context
var driver neo4j.DriverWithContext
var SecureMode bool
var port = flag.Int("port", 3333, "the port to serve on")

// Only edit to insecure type datasets
var datasetNumberWriting = "200_i"
var workloadType = "ff/"

var csvPath = "../../graphs/" + workloadType + "latencyDatabase_" + datasetNumberWriting + ".csv"
var csvFile *os.File

// should implement the interface myPkgName.ServiceServer
type serviceServer struct {
	// type embedded to comply with Google lib
	service.UnimplementedServiceServer
}

func checkIntegrity() bool {
	blockchainResults := getByRangeAsset()
	if blockchainResults != nil {
		//Query Primary database
		resultNodes, err := neo4j.ExecuteQuery(ctx, driver,
			`MATCH (n) RETURN n`,
			nil,
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase("neo4j"))
		if err != nil {
			panic(err)
		}
		resultRelationships, err := neo4j.ExecuteQuery(ctx, driver,
			`MATCH p=()-->() RETURN p`,
			nil,
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase("neo4j"))
		if err != nil {
			panic(err)
		}
		//Updating Verifier
		fmt.Println("Updating Verifier database")
		for _, element := range blockchainResults {
			_, err := neo4j.ExecuteQuery(ctxV, driverV,
				element.Asset,
				nil,
				neo4j.EagerResultTransformer,
				neo4j.ExecuteQueryWithDatabase("neo4j"))
			if err != nil {
				panic(err)
			}
		}

		//Querying Verifier database
		resultVNodes, err := neo4j.ExecuteQuery(ctxV, driverV,
			`MATCH (n) RETURN n`,
			nil,
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase("neo4j"))
		if err != nil {
			panic(err)
		}
		resultVRelationships, err := neo4j.ExecuteQuery(ctxV, driverV,
			`MATCH p=()-->() RETURN p`,
			nil,
			neo4j.EagerResultTransformer,
			neo4j.ExecuteQueryWithDatabase("neo4j"))
		if err != nil {
			panic(err)
		}
		//Confirming if both databases possess the same data state
		if !checkAssets(resultNodes, resultVNodes) {
			return false
		}
		if !checkAssets(resultRelationships, resultVRelationships) {
			return false
		}
	}
	return true
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

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connection established.")

	return ctx, driver
}

func disconnectDatabase() {
	defer func(driver neo4j.DriverWithContext, ctx context2.Context) {
		err := driver.Close(ctx)
		if err != nil {
			fmt.Printf("Error: ")
			fmt.Println(err.Error())
		}
	}(driver, ctx)
}

func writeLatency(start time.Time, csvFile *os.File) {
	elapsed := time.Since(start).Milliseconds()
	_, err := fmt.Fprintf(csvFile, "%d\n", elapsed)
	if err != nil {
		panic(err)
	}
}

// WriteDatabase Create a node representing a person named Alice
func (m *serviceServer) WriteDatabase(ctx_c context.Context, request *service.WriteDatabaseRequest) (*service.WriteDatabaseResponse, error) {
	var start time.Time
	if !SecureMode {
		start = time.Now()
	}
	result, err := neo4j.ExecuteQuery(ctx, driver,
		request.Value,
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		panic(err)
	}
	if SecureMode {
		createAsset(request.Value)
	} else {
		writeLatency(start, csvFile)
	}
	fmt.Printf("Created %v nodes in %+v.\n",
		result.Summary.Counters().NodesCreated(),
		result.Summary.ResultAvailableAfter())
	return &service.WriteDatabaseResponse{Value: "WRITE SUCCESS"}, nil
}

func (m *serviceServer) ReadDatabase(ctx_c context.Context, request *service.ReadDatabaseRequest) (*service.ReadDatabaseResponse, error) {
	var start time.Time
	if SecureMode {
		if !checkIntegrity() {
			return &service.ReadDatabaseResponse{Value: "READ NOT SUCCESS: DATABASE COMPROMISED"}, nil
		}
	} else {
		start = time.Now()
	}
	result, err := neo4j.ExecuteQuery(ctx, driver,
		request.Value,
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if !SecureMode {
		writeLatency(start, csvFile)
	}
	resultJson, _ := json.Marshal(result.Records)
	return &service.ReadDatabaseResponse{Value: string(resultJson)}, nil
}

// UpdateDatabase Update node Alice to add an age property
func (m *serviceServer) UpdateDatabase(ctx_c context.Context, request *service.UpdateDatabaseRequest) (*service.UpdateDatabaseResponse, error) {
	result, err := neo4j.ExecuteQuery(ctx, driver,
		request.Value,
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		panic(err)
	}
	resultJson, _ := json.Marshal(result.Records)
	return &service.UpdateDatabaseResponse{Value: string(resultJson)}, nil
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
	caPem, err := ioutil.ReadFile("../cert/ca-cert.pem")
	if err != nil {
		log.Fatal(err)
	}

	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		log.Fatal(err)
	}

	// read server cert & key
	serverCert, err := tls.LoadX509KeyPair("../cert/server-cert.pem", "../cert/server-key.pem")
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
	//Checking the input flags when the server starts. More specifically the secure mode, which implies communication with the Hyperledger Fabric blockchain
	if len(os.Args) < 2 {
		fmt.Println("Missing parameter, provide Query Mode flag!")
		return
	} else if os.Args[1] == "-i" {
		fmt.Println("INSECURE MODE")
		SecureMode = false
	} else if os.Args[1] == "-s" {
		fmt.Println("SECURE MODE")
		SecureMode = true
	} else {
		fmt.Println("Invalid parameter, provide Query Mode flag correctly!")
		return
	}

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

	if SecureMode {
		//Create a blockchain connection
		blockchainConnectionStartup()
		connectVerifier()
	} else {
		csvFile, err = os.Create(csvPath)
		_, err = fmt.Fprintf(csvFile, "Latency\n")
	}

	//Handles CTRL^C signal to execute a graceful exit by closing the database connection
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		s := <-c
		if s == syscall.SIGINT {
			fmt.Print(" ")
			fmt.Println("Detected")
			fmt.Println("Closing databases & blockchain connections...")
			disconnectDatabase()
			if SecureMode {
				disconnectVerifier()
				err := gw.Close()
				if err != nil {
					return
				}
				err = clientConnection.Close()
				if err != nil {
					return
				}
				err = csvFile.Close()
				if err != nil {
					log.Fatal(err)
				}
			}
			fmt.Println("Connections closed with success")
			os.Exit(0)
		}
	}()
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
