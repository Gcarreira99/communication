package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

// NodeLabels contains the labels of nodes
var NodeLabels = []string{"Hashtag", "Link", "Source", "Tweet", "User"}

// RelationshipTypes defines the relationship types and the corresponding from-to labels
var RelationshipTypes = map[string]map[string][]string{
	"CONTAINS": {
		"Tweet": {"Source"},
	},
	"FOLLOWS": {
		"User": {"User"},
	},
	"MENTIONS": {
		"Tweet": {"User"},
	},
	"POSTS": {
		"User": {"Tweet"},
	},
	"REPLY_TO": {
		"Tweet": {"Tweet"},
	},
	"RETWEETS": {
		"Tweet": {"Tweet"},
	},
	"TAGS": {
		"Tweet": {"Hashtag"},
	},
	"USING": {
		"Tweet": {"Source"},
	},
}

// NodeProperties contains the properties of nodes
var NodeProperties = make(map[string]map[string]string)

func main() {
	// Open file for writing
	file, err := os.Create("generated_queries_400.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Redirect output to file
	fmt.Println("Generated queries are written into generated_queries_25.txt")
	os.Stdout = file

	// Generate CREATE queries for nodes
	nodeQueries := make([]string, 0, len(NodeLabels))
	for _, label := range NodeLabels {
		properties := GenerateRandomProperties(label)
		NodeProperties[label] = properties
		createQuery := GenerateCreateQuery(label, properties)
		nodeQueries = append(nodeQueries, createQuery)
	}

	// Generate relationship creation queries
	relationshipQueries := make([]string, 0, 25) // We'll generate 25 relationships initially
	for relationshipType, fromToLabels := range RelationshipTypes {
		for fromLabel, toLabels := range fromToLabels {
			for _, toLabel := range toLabels {
				// Generate relationship queries with proper conditions on nodes' properties and values
				fromProperty, fromValue := getRandomPropertyAndValue(fromLabel)
				toProperty, toValue := getRandomPropertyAndValue(toLabel)
				relationshipQuery := fmt.Sprintf("MATCH (n:%s {%s: '%s'}), (m:%s {%s: '%s'}) WHERE n.%s = '%s' AND m.%s = '%s' CREATE (n)-[:%s]->(m)", fromLabel, fromProperty, fromValue, toLabel, toProperty, toValue, fromProperty, fromValue, toProperty, toValue, relationshipType)
				if relationshipQuery != "" {
					relationshipQueries = append(relationshipQueries, relationshipQuery)
				}
			}
		}
	}

	// Shuffle node and relationship queries separately
	rand.Shuffle(len(nodeQueries), func(i, j int) {
		nodeQueries[i], nodeQueries[j] = nodeQueries[j], nodeQueries[i]
	})
	rand.Shuffle(len(relationshipQueries), func(i, j int) {
		relationshipQueries[i], relationshipQueries[j] = relationshipQueries[j], relationshipQueries[i]
	})

	// Determine the number of additional node queries needed to reach 200
	numAdditionalNodeQueries := 200 - len(nodeQueries)
	if numAdditionalNodeQueries > 0 {
		for i := 0; i < numAdditionalNodeQueries; i++ {
			label := NodeLabels[rand.Intn(len(NodeLabels))]
			properties := GenerateRandomProperties(label)
			NodeProperties[label] = properties
			createQuery := GenerateCreateQuery(label, properties)
			nodeQueries = append(nodeQueries, createQuery)
		}
	}

	// Determine the number of additional relationship queries needed to reach 200
	numAdditionalRelationshipQueries := 200 - len(relationshipQueries)
	if numAdditionalRelationshipQueries > 0 {
		for i := 0; i < numAdditionalRelationshipQueries; i++ {
			relationshipType := ""
			for rType := range RelationshipTypes {
				relationshipType = rType
				break
			}
			var fromLabel string
			for fl := range RelationshipTypes[relationshipType] {
				fromLabel = fl
				break
			}
			toLabels := RelationshipTypes[relationshipType][fromLabel]
			toLabel := toLabels[rand.Intn(len(toLabels))]
			// Generate relationship queries with proper conditions on nodes' properties and values
			fromProperty, fromValue := getRandomPropertyAndValue(fromLabel)
			toProperty, toValue := getRandomPropertyAndValue(toLabel)
			relationshipQuery := fmt.Sprintf("MATCH (n:%s {%s: '%s'}), (m:%s {%s: '%s'}) WHERE n.%s = '%s' AND m.%s = '%s' CREATE (n)-[:%s]->(m)", fromLabel, fromProperty, fromValue, toLabel, toProperty, toValue, fromProperty, fromValue, toProperty, toValue, relationshipType)
			if relationshipQuery != "" {
				relationshipQueries = append(relationshipQueries, relationshipQuery)
			}
		}
	}

	// Combine node and relationship queries to get 200 queries
	allQueries := make([]string, 0, 200)
	allQueries = append(allQueries, nodeQueries...)
	allQueries = append(allQueries, relationshipQueries...)

	// Output the queries
	for _, query := range allQueries {
		fmt.Println(query)
	}
}

// Function to generate a random property and value for a given node label
func getRandomPropertyAndValue(label string) (string, string) {
	switch label {
	case "Hashtag", "Source":
		return "name", getRandomValue()
	case "Link":
		return "url", getRandomValue()
	case "Tweet":
		return "id_str", strconv.Itoa(rand.Intn(1000000)) // Assuming the id_str is only numeric
	case "User":
		properties := []string{"followers", "screen_name", "following", "name", "profile_image_url", "location", "url"}
		property := properties[rand.Intn(len(properties))]
		return property, getRandomValue()
	default:
		return "", ""
	}
}

// Function to generate a random value
func getRandomValue() string {
	return strconv.Itoa(rand.Intn(1000000))
}

// Function to generate random properties for a given node label
func GenerateRandomProperties(label string) map[string]string {
	properties := make(map[string]string)
	switch label {
	case "Hashtag", "Source":
		properties["name"] = getRandomValue()
	case "Link":
		properties["url"] = "http://example.com/" + getRandomValue()
	case "Tweet":
		properties["favorites"] = getRandomValue()
		properties["import_method"] = "user"
		properties["id_str"] = getRandomValue()
		properties["created_at"] = "2022-01-01T00:00:00Z"
		properties["text"] = "This is a sample tweet."
	case "User":
		properties["followers"] = getRandomValue()
		properties["screen_name"] = "user" + getRandomValue()
		properties["following"] = getRandomValue()
		properties["name"] = "User " + getRandomValue()
		properties["profile_image_url"] = "http://example.com/image" + getRandomValue() + ".jpg"
		properties["location"] = "Location " + getRandomValue()
		properties["url"] = "http://example.com/user" + getRandomValue()
	}
	return properties
}

// Function to generate a CREATE query for a node with given properties
func GenerateCreateQuery(label string, properties map[string]string) string {
	query := "CREATE (m:" + label + "{"
	for key, value := range properties {
		query += key + ": '" + value + "', "
	}
	query = query[:len(query)-2] + "})"
	return query
}
