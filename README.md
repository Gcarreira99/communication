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
.
├── server
│   ├── server.go
│   └── ledger_bridge.go
├── service
│   └── common.proto
├── client
│   └── client.go
├── scripts
│   ├── cert_generator.sh
│   └── database_startup.sh
├── chaincode
│   └── smartcontract.go
├── cert
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
./network.sh deployCC -ccn basic -ccp ~/project/chaincode/ -ccl go
```
3. [OPTIONAL] Open another terminal and run the following command to monitor the network:
```shellscript
./monitordocker.sh
```

**_NOTE:_** **`~/project/chaincode/`** is the chaincode folder path.

### Deploy Neo4j Database
1. Run the **`database_startup.sh`** to deploy a Neo4j database in a container.
2. Send the dump file to the docker container:
```shellscript
docker cp ./twitter-v2-50.dump "docker_id":/var/lib/neo4j/twitter-v2-50.dump
```
3. Restore database from the dump file:
```shellscript
neo4j-admin database load --from-path=/var/lib/neo4j/ twitter-v2-50 --overwrite-destination=true --verbose
```
4. Overwrite the original info from **`neo4j`** database with **`twitter-v2-50`** database info:
```shellscript
rm -rf databases/neo4j/*
cp -r --verbose databases/twitter-v2-50/* databases/neo4j
cp -r --verbose transactions/twitter-v2-50/* transactions/neo4j
```

### Generate TLS Information
- Run the script:
```shellscript
./cert_generator.sh
```

### Run proto file for generation
- Run this command inside the main folder from the project:
```shellscript
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative service/common.proto
```
### Run System
1. To run the system's server is necessary the following commands:
```shellscript
cd server
go build
sudo ./server
```
2. To run the system's client is necessary the following commands:
```shellscript
cd client
go build
sudo ./client
```
**_NOTE:_** Both client and server need to run in **`sudo`** to access the TLS information.