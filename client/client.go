package main

import (
	"bufio"
	pb "communication/service"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var FileMode bool
var datasetNumberReading = "200_ff"

var datasetNumberWriting = "200_i"
var workloadType = "ff/"

func sendRequest(client pb.ServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	answer, err := client.Create(ctx, &pb.CreateRequest{From: "A", To: "B", Total: nil})
	if err != nil {
		log.Fatalf("client.Create failed: %v", err)
	}
	log.Println(answer)
}

func writeRequest(client pb.ServiceClient, request string) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()
	answer, err := client.WriteDatabase(ctx, &pb.WriteDatabaseRequest{Value: request})
	if err != nil {
		log.Fatalf("client.Create failed: %v", err)
	}
	log.Println(answer)
}

func readRequest(client pb.ServiceClient, query string) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
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

func fileReading(client pb.ServiceClient) *os.File {
	queryPath := "../scripts/generated_queries_" + datasetNumberReading + ".txt"
	statsPath := "../../graphs/" + workloadType + "stats_" + datasetNumberWriting + ".csv"
	data, err := os.Open(queryPath)
	if err != nil {
		log.Fatal(err)
	}
	// Create a new Scanner for the file.
	scanner := bufio.NewScanner(data)
	//Csv Writer
	csvFile, err := os.Create(statsPath)
	_, err = fmt.Fprintf(csvFile, "Dataset,Latency,Usage\n")
	if err != nil {
		log.Fatal(err)
	}
	// Loop over all lines in the file and print them.
	for scanner.Scan() {
		line := scanner.Text()
		start := time.Now()
		if strings.Contains(line, "CREATE") || strings.Contains(line, "DELETE") || strings.Contains(line, "SET") {
			writeRequest(client, line)
		} else {
			readRequest(client, line)
		}
		elapsed := time.Since(start).Milliseconds()
		writeStats(elapsed, csvFile)
	}
	return csvFile
}

func writeStats(elapsed int64, csvFile *os.File) {
	datasetName := "dataset_" + datasetNumberWriting

	size, err := DirSize("../")
	if err != nil {
		panic(err)
	}
	_, err = fmt.Fprintf(csvFile, "%s,%d,%d\n", datasetName, elapsed, size)
	if err != nil {
		panic(err)
	}
}

func writeThroughput(elapsed int64) {
	throughputPath := "../../graphs/" + workloadType + "throughput_" + datasetNumberWriting + ".csv"
	datasetName := "dataset_" + datasetNumberWriting
	csvFile, err := os.Create(throughputPath)
	_, err = fmt.Fprintf(csvFile, "x,y\n")
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	_, err = fmt.Fprintf(csvFile, "%s,%d\n", datasetName, elapsed)
	err = csvFile.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing parameter, provide Reading Mode flag!")
		return
	}
	//datasetNumberReading = os.Args[3]
	//datasetNumberWriting = os.Args[3] + "_" + os.Args[2]
	if os.Args[1] == "-f" {
		FileMode = true
	} else if os.Args[1] == "-c" {
		FileMode = false
	} else {
		fmt.Println("Invalid parameter, provide Query Mode flag correctly!")
		return
	}

	// Create tls based credential.
	creds := loadCertificates()

	conn, err := grpc.Dial("0.0.0.0:3333", grpc.WithTransportCredentials(creds))
	if err != nil {
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
	if FileMode == true {
		start := time.Now()
		r := new(big.Int)
		fmt.Println(r.Binomial(1000, 10))

		fmt.Printf("File Mode: reading file 'generated_queries_%s.txt'", datasetNumberReading)
		csvFile := fileReading(client)
		fmt.Println("Success reading file!")
		fmt.Println("Closing...")
		err := csvFile.Close()
		if err != nil {
			return
		}
		elapsed := time.Since(start).Milliseconds()
		writeThroughput(elapsed)
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
