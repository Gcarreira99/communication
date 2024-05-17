package main

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"golang.org/x/net/context"
)

func connectDatabase() (context.Context, neo4j.DriverWithContext) {
	ctx := context.Background()

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

func tampering(ctx context.Context, session neo4j.SessionWithContext) {
	_, err := session.Run(ctx, `MATCH (n:User) WHERE n.screen_name='Batmanandsuper1' DETACH DELETE n`, nil)
	if err != nil {
		panic(err)
	}
	_, err = session.Run(ctx, `MERGE (n:User {name: 'Jakub Bares'}) SET n = {name: 'John', location: 'Lisbon', following: 1092}
RETURN n`, nil)
	if err != nil {
		panic(err)
	}
	_, err = session.Run(ctx, `CREATE (m:Hashtag{name: 'ATTACKER'})`, nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx, driver := connectDatabase()
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	tampering(ctx, session)
}
