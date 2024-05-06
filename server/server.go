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
)

var ctx context.Context
var driver neo4j.DriverWithContext
var session neo4j.SessionWithContext
var SecureMode bool
var port = flag.Int("port", 3333, "the port to serve on")

// should implement the interface myPkgName.ServiceServer
type serviceServer struct {
	// type embedded to comply with Google lib
	service.UnimplementedServiceServer
}

func checkIntegrity() bool {
	result, err := session.Run(ctx, `MATCH (n)-[r]->(m) RETURN n,r,m`, nil)
	if err != nil {
		// Handle error
		fmt.Println("Error querying database:", err)
		panic(err)
	}

	//Querying Verifier
	results := getAllAssets()
	if results != nil {
		//Creating/Updating Verifier database
		for _, element := range results {
			fmt.Printf("TRANSACTION: %s\n", element.Asset)
			_, errV := sessionV.Run(ctxV, element.Asset, nil)
			if errV != nil {
				// Handle error
				fmt.Println("CREATE ASSET: Error querying verifier:", errV)
				panic(err)
			}
		}
	}
	//Querying the all Verifier database
	resultV, errV := sessionV.Run(ctxV, `MATCH (n)-[r]->(m) RETURN n,r,m`, nil)
	if errV != nil {
		// Handle error
		fmt.Println("Error querying verifier:", errV)
		panic(err)
	}
	//Confirming if both databases possess the same data state
	var state bool
	if result.Record() == resultV.Record() {
		fmt.Println("TRUE: same state!")
		state = true
	} else {
		fmt.Println("FALSE: different state!")
		state = false
	}
	return state
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

// WriteDatabase Create a node representing a person named Alice
func (m *serviceServer) WriteDatabase(ctx_c context.Context, request *service.WriteDatabaseRequest) (*service.WriteDatabaseResponse, error) {
	/*
		if SecureMode {
			if !checkIntegrity() {
				return &service.WriteDatabaseResponse{Value: "WRITE NOT SUCCESS: DATABASE COMPROMISED"}, nil
			}
		}
	*/
	result, err := neo4j.ExecuteQuery(ctx, driver,
		request.Value,
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		panic(err)
	}
	/*
		if SecureMode {
			createAsset(request.Value)
		}
	*/
	fmt.Printf("Created %v nodes in %+v.\n",
		result.Summary.Counters().NodesCreated(),
		result.Summary.ResultAvailableAfter())
	return &service.WriteDatabaseResponse{Value: "WRITE SUCCESS"}, nil
}

// ReadDatabase Retrieve all Person nodes
func (m *serviceServer) ReadDatabase(ctx_c context.Context, request *service.ReadDatabaseRequest) (*service.ReadDatabaseResponse, error) {
	if SecureMode {
		if !checkIntegrity() {
			return &service.ReadDatabaseResponse{Value: "READ NOT SUCCESS: DATABASE COMPROMISED"}, nil
		}
		if request.Value == "ALL" {
			getAllAssets()
			return &service.ReadDatabaseResponse{Value: "READ SUCCESS"}, nil
		}
	}
	result, err := neo4j.ExecuteQuery(ctx, driver,
		request.Value,
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		panic(err)
	}
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
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
	session = driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})

	if SecureMode {
		//Create a blockchain connection
		blockchainConnectionStartup()
		connectVerifier()
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
			disconnectVerifier()
			if SecureMode {
				err := gw.Close()
				if err != nil {
					return
				}
				err = clientConnection.Close()
				if err != nil {
					return
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
