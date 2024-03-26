# System Overall Setup

## Used Environment
- Hyperledger Fabric [test-network]
- WSL
- Golang
- Neo4j
- Docker Desktop

## Installation
- Follow the installation instructions from the [Hyperledger Fabric guide](https://hyperledger-fabric.readthedocs.io/en/release-2.5/getting_started.html).

## File Structure
The system file structure is the following:
```
communication/
├── server/
│   ├── server.go
│   └── ledger_bridge.go
├── service/
│   └── common.proto
├── client/
│   └── client.go
├── scripts/
│   ├── cert_generator.sh
│   └── database_startup.sh
├── chaincode/
│   └── smartcontract.go
├── cert/
│   ├── server-ext.conf
│   └── client-ext.conf
│   
└── README.md
```

## Usage
1. Deploy Hyperledger Fabric network, more precisely **`test-network`**.
2. Deploy the Neo4j database
3. Generate the TLS information to enable communication between the client and server.
4. Run the proto file to generate the necessary gRPC functions for the client-server communication.
5. Run the system, both client and server

### Deploy **`test-network`**
1. Go to the **`fabric-samples`** repository the was downloaded during the Installation section. After that go to the **`test-network`** folder.
2. Run following commands in one terminal, one after the other finishes:
```shellscript
./network.sh down
./network.sh up createChannel -c mychannel -ca
./network.sh deployCC -ccn basic -ccp path/to/chaincode/folder/ -ccl go
```
3. [OPTIONAL] Open another terminal and run the following command to monitor the network:
```shellscript
./monitordocker.sh
```

### Deploy Neo4j Database
1. Run the database startup script to deploy a Neo4j database inside the docker container.
```shellscript
cd scripts
./database_startup.sh
```
2. Download the [dump file](https://github.com/neo4j-graph-examples/twitter-v2/blob/main/data/twitter-v2-50.dump) that was used in this project.
3. To continue the guide is needed the ID of the docker where the Neo4j database is running. To access the ID, run the following command:
```shellscript
docker ps
```
4. Send the dump file to the docker container:
```shellscript
docker cp ./twitter-v2-50.dump docker_id:/var/lib/neo4j/twitter-v2-50.dump
```
5. Restore database from the dump file:
```shellscript
neo4j-admin database load --from-path=/var/lib/neo4j/ twitter-v2-50 --overwrite-destination=true --verbose
```
6. Overwrite the original info from **`neo4j`** database with **`twitter-v2-50`** database info:
```shellscript
rm -rf databases/neo4j/*
cp -r --verbose databases/twitter-v2-50/* databases/neo4j
cp -r --verbose transactions/twitter-v2-50/* transactions/neo4j
```
**_NOTE:_** There is a probability that eventually the Neo4j container will crash and present some problems when booting. In that case remove the container and run again the script. 

### Generate TLS Information
- Run the script:
```shellscript
cd scripts
./cert_generator.sh
```
- After running the script is possible to access the certificates and keys inside the **`cert`** folder.

### Run proto file for generation
- Run this command inside the main folder from the project:
```shellscript
cd service
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative service/common.proto
```
### Run System
- To run the system's server is necessary the following commands:
```shellscript
cd server
go build
./server
```
- To run the system's client is necessary the following commands:
```shellscript
cd client
go build
./client
```
- To close the server is press **`Ctrl^C`**.
- To close the client is only needed to use the client's shell and to type the **`exit`** command.