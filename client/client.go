package main

import (
	"bufio"
	pb "communication/service"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

var (
	serverAddr = flag.String("addr", "localhost:3333", "The server address in the format of host:port")
)

var FILE_MODE bool

func sendRequest(client pb.ServiceClient) {
	//log.Printf("Looking for features within %v", request)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	answer, err := client.Create(ctx, &pb.CreateRequest{From: "A", To: "B", Total: nil})
	if err != nil {
		log.Fatalf("client.Create failed: %v", err)
	}
	log.Println(answer)
}

func writeRequest(client pb.ServiceClient, request string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	answer, err := client.WriteDatabase(ctx, &pb.WriteDatabaseRequest{Value: request})
	if err != nil {
		log.Fatalf("client.Create failed: %v", err)
	}
	log.Println(answer)
}

func readRequest(client pb.ServiceClient, query string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	answer, err := client.ReadDatabase(ctx, &pb.ReadDatabaseRequest{Value: query})
	if err != nil {
		log.Fatalf("client.Create failed: %v", err)
	}
	log.Println(answer.Value)
}

func updateRequest(client pb.ServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	answer, err := client.UpdateDatabase(ctx, &pb.UpdateDatabaseRequest{Value: "Alice"})
	if err != nil {
		log.Fatalf("client.Create failed: %v", err)
	}
	log.Println(answer)
}

func deleteRequest(client pb.ServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	answer, err := client.DeleteDatabase(ctx, &pb.DeleteDatabaseRequest{Value: "Alice"})
	if err != nil {
		log.Fatalf("client.Create failed: %v", err)
	}
	log.Println(answer)
}

func loadCertificates() credentials.TransportCredentials {
	// read ca's cert
	caCert, err := ioutil.ReadFile("../cert/ca-cert.pem")
	if err != nil {
		log.Fatal(caCert)
	}
	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatal(err)
	}

	//read client cert
	clientCert, err := tls.LoadX509KeyPair("../cert/client-cert.pem", "../cert/client-key.pem")
	if err != nil {
		log.Fatal(err)
	}

	// set config of tls credential
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	tlsCredential := credentials.NewTLS(config)

	return tlsCredential

}

func commandOptions(reader *bufio.Reader) string {
	fmt.Print("-->> ")
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1)
	return text
}

func fileReading(client pb.ServiceClient) {
	data, err := os.Open("../scripts/generated_queries.txt")
	if err != nil {
		log.Fatal(err)
	}
	// Create a new Scanner for the file.
	scanner := bufio.NewScanner(data)
	// Loop over all lines in the file and print them.
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "CREATE") {
			writeRequest(client, line)
		} else {
			readRequest(client, line)
		}
	}
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Missing parameter, provide Reading Mode flag!")
		return
	}
	if os.Args[1] == "-f" {
		FILE_MODE = true
	} else if os.Args[1] == "-c" {
		FILE_MODE = false
	} else {
		fmt.Println("Invalid parameter, provide Query Mode flag correctly!")
		return
	}

	// Create tls based credential.
	creds := loadCertificates()

	//conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(creds))
	conn, err := grpc.Dial("0.0.0.0:3333", grpc.WithTransportCredentials(creds))
	if err != nil {
		//log.Fatalf("fail to dial: %v", err)
		log.Fatal(err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)
	client := pb.NewServiceClient(conn)
	sendRequest(client)

	//Once the file mode is active the client exits when finishing reading the input file
	if FILE_MODE == true {
		start := time.Now()
		r := new(big.Int)
		fmt.Println(r.Binomial(1000, 10))

		fmt.Println("File Mode: reading file 'generated_queries.txt'")
		fileReading(client)
		fmt.Println("Success reading file!")
		fmt.Println("Closing...")

		elapsed := time.Since(start)
		log.Printf("Binomial took %s", elapsed)
		os.Exit(0)
	}

	//If file mode is not active the client accepts input commands from the command line
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Communication Shell")
	fmt.Println("-------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		switch text {
		case "write":
			fmt.Println("write command")
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)
			writeRequest(client, text)
		case "read":
			fmt.Println("read command")
			readRequest(client, commandOptions(reader))
		case "update":
			fmt.Println("update command")
			updateRequest(client)
		case "delete":
			fmt.Println("delete command")
			deleteRequest(client)
		case "exit":
			fmt.Println("Closing...")
			os.Exit(0)
		default:
			fmt.Println("wrong command")
		}
	}
}
