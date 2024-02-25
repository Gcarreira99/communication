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
	"os"
	"strings"
	"time"
)

var (
	serverAddr = flag.String("addr", "localhost:3333", "The server address in the format of host:port")
)

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

func writeRequest(client pb.ServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	answer, err := client.WriteDatabase(ctx, &pb.WriteDatabaseRequest{Value: "Alice"})
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
	log.Println(answer)
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
	caCert, err := ioutil.ReadFile("/keys/cd/ca-cert.pem")
	if err != nil {
		log.Fatal(caCert)
	}
	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatal(err)
	}

	//read client cert
	clientCert, err := tls.LoadX509KeyPair("/keys/cd/client-cert.pem", "/keys/cd/client-key.pem")
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

func main() {

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

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Communication Shell")
	fmt.Println("-------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		if strings.Compare("hi", text) == 0 {
			fmt.Println("Hello, Yourself")
		}

		switch text {
		case "write":
			fmt.Println("write command")
			writeRequest(client)
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
