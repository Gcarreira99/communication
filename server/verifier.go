package main

import (
	context2 "context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"golang.org/x/net/context"
)

var ctxV context.Context
var driverV neo4j.DriverWithContext
var sessionV neo4j.SessionWithContext

func connectVerifier() {
	var err error
	ctxV = context.Background()
	// URI examples: "neo4j://localhost", "neo4j+s://xxx.databases.neo4j.io"
	dbUri := "neo4j://localhost:7688"
	dbUser := "neo4j"
	dbPassword := "password"
	driverV, err = neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""))
	if err != nil {
		panic(err)
	}

	err = driverV.VerifyConnectivity(ctxV)
	if err != nil {
		panic(err)
	}
	sessionV = driverV.NewSession(ctxV, neo4j.SessionConfig{DatabaseName: "neo4j"})
	fmt.Println("Verifier connection established.")
}

func disconnectVerifier() {
	defer func(driverV neo4j.DriverWithContext, ctxV context2.Context) {
		err := driverV.Close(ctxV)
		if err != nil {
			fmt.Printf("Error: ")
			fmt.Println(err.Error())
		}
	}(driverV, ctxV)
}

// Don't use, but don't delete
func startupData() {
	//Query all nodes
	result, err := neo4j.ExecuteQuery(ctx, driver,
		"MATCH (n)-[r]->(m) RETURN n,r,m LIMIT 25",
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Records[0].Values[0])

	for _, record := range result.Records {
		// Extract nodes n and m and relationship r from the record
		// Extract nodes
		nodeN, err := extractNode(record.Values[0])
		if err != nil {
			fmt.Println("Error extracting node n:", err)
			return
		}

		relationshipTemp := record.Values[1]
		relationshipR := relationshipTemp.(dbtype.Relationship)

		nodeM, err := extractNode(record.Values[2])
		if err != nil {
			fmt.Println("Error extracting node m:", err)
			return
		}

		// Assuming you want to create nodes and relationships with the same properties and labels/types
		// Construct Cypher queries to create nodes and relationships in the new database
		// Construct Cypher queries to create nodes and relationships in the new database
		cypherCreateNodeN := fmt.Sprintf("CREATE (n:%s%s)", labelsString(nodeN.Labels), propertiesString(nodeN.Props))
		cypherCreateNodeM := fmt.Sprintf("CREATE (m:%s%s)", labelsString(nodeM.Labels), propertiesString(nodeM.Props))
		cypherCreateRelationship := fmt.Sprintf("MATCH (n), (m) WHERE n.id = '%v' AND m.id = '%v' CREATE (n)-[:%s {props}]->(m)", nodeN.ElementId, nodeM.ElementId, relationshipR.Type)

		// Execute the Cypher queries to populate the new database
		_, err = sessionV.Run(ctxV, cypherCreateNodeN, map[string]interface{}{"props": nodeN.Props})
		if err != nil {
			// Handle error
			fmt.Println("Error creating node n:", err)
			return
		}

		_, err = sessionV.Run(ctxV, cypherCreateNodeM, map[string]interface{}{"props": nodeM.Props})
		if err != nil {
			// Handle error
			fmt.Println("Error creating node m:", err)
			return
		}

		_, err = sessionV.Run(ctxV, cypherCreateRelationship, map[string]interface{}{"props": relationshipR.Props})
		if err != nil {
			// Handle error
			fmt.Println("Error creating relationship:", err)
			return
		}
	}
}

// Don't use, but don't delete
func extractNode(nodeInterface interface{}) (neo4j.Node, error) {
	node, ok := nodeInterface.(neo4j.Node)
	if !ok {
		return neo4j.Node{}, fmt.Errorf("could not extract node")
	}
	return node, nil
}

// Don't use, but don't delete
func labelsString(labels []string) string {
	returnString := labels[0]
	for i := 1; i < len(labels); i++ {
		returnString = returnString + ":" + labels[i]
	}
	return returnString
}

// Don't use, but don't delete
func propertiesString(properties map[string]interface{}) string {
	props := "{"
	for key, value := range properties {
		props += key + ": '" + fmt.Sprintf("%v", value) + "', "
	}
	props += "}"
	props = props[0 : len(props)-3]
	props += "}"
	return props
}
