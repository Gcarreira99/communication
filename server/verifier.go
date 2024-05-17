package main

import (
	context2 "context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"golang.org/x/net/context"
	"reflect"
)

var ctxV context.Context
var driverV neo4j.DriverWithContext

func connectVerifier() {
	var err error
	ctxV = context.Background()
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

func assertEq(variable1 any, variable2 any) bool {
	return reflect.DeepEqual(variable1, variable2)
}

func checkAssets(result *neo4j.EagerResult, resultV *neo4j.EagerResult) bool {
	records := result.Records
	recordsV := resultV.Records
	if len(records) == len(recordsV) {
		for i := range records {
			values := records[i].Values
			valuesV := recordsV[i].Values
			switch interface{}(values).(type) {
			case dbtype.Node:
				node := values[0].(dbtype.Node)
				nodeV := valuesV[0].(dbtype.Node)
				if !assertEq(node.Labels, nodeV.Labels) || !assertEq(node.GetProperties(), nodeV.GetProperties()) {
					return false
				}
			case dbtype.Path:
				relationship := values[0].(dbtype.Path)
				relationshipV := valuesV[0].(dbtype.Path)
				if !assertEq(relationship.Nodes, relationshipV.Nodes) || !assertEq(relationship.Relationships, relationshipV.Relationships) {
					return false
				}
			}
		}
		return true
	}
	return false
}
