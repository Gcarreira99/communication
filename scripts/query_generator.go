package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NodeLabels represents the available node labels
var NodeLabels = []string{"Hashtag", "Link", "Source", "Tweet", "User"}

// RelationshipTypes represents the available relationship types
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

// PropertyKeys represents the available property keys
var PropertyKeys = map[string][]string{
	"Hashtag": {"name"},
	"Link":    {"url"},
	"Source":  {"name"},
	"Tweet": {
		"favorites", "import_method", "id_str", "created_at", "id", "text",
	},
	"User": {
		"followers", "screen_name", "following", "name", "profile_image_url", "location", "url",
	},
}

// GenerateRandomProperties generates random properties based on node label
func GenerateRandomProperties(label string) map[string]interface{} {
	properties := make(map[string]interface{})
	rand.Seed(time.Now().UnixNano())
	keys := PropertyKeys[label]
	for _, key := range keys {
		// Generate random value based on key
		var value interface{}
		switch key {
		case "favorites":
			value = rand.Intn(100)
		case "import_method":
			value = "user"
		case "id_str":
			value = generateNumericID()
		case "created_at":
			value = time.Now().Format(time.RFC3339)
		case "id":
			value = generateNumericID()
		case "text":
			value = "Random text " + uuid.New().String()[:4]
		case "followers":
			value = rand.Intn(100000)
		case "screen_name":
			value = "User" + uuid.New().String()[:4]
		case "following":
			value = rand.Intn(100000)
		case "name":
			value = "Name_" + uuid.New().String()[:4]
		case "profile_image_url":
			value = "http://random.url/" + uuid.New().String()[:4] + ".jpg"
		case "location":
			value = "Location_" + uuid.New().String()[:4]
		case "url":
			value = "http://random.url/" + uuid.New().String()[:4]
		default:
			value = "Unknown"
		}
		properties[key] = value
	}
	return properties
}

// GenerateCreateQuery generates a CREATE query
func GenerateCreateQuery(label string, properties map[string]interface{}) string {
	var props []string
	for key, value := range properties {
		switch v := value.(type) {
		case string:
			props = append(props, fmt.Sprintf(`%s: '%s'`, key, strings.ReplaceAll(v, `"`, `'`)))
		case int:
			props = append(props, fmt.Sprintf(`%s: %d`, key, v))
		}
	}
	return fmt.Sprintf(`CREATE (m:%s {%s})`, label, strings.Join(props, ", "))
}

// GenerateRelationshipQuery generates a relationship creation query
func GenerateRelationshipQuery(relationshipType, fromLabel, toLabel string) string {
	conditions := make([]string, 0)

	switch {
	case fromLabel == "User" && toLabel == "User":
		conditions = append(conditions, fmt.Sprintf("n.url = 'http://random.url/%s'", uuid.New().String()[:4]))
		conditions = append(conditions, fmt.Sprintf("m.url = 'http://random.url/%s'", uuid.New().String()[:4]))
	case fromLabel == "Tweet" && toLabel == "Tweet":
		conditions = append(conditions, fmt.Sprintf("n.id_str = '%s'", generateNumericID()))
		conditions = append(conditions, fmt.Sprintf("m.id_str = '%s'", generateNumericID()))
	case fromLabel == "User" && toLabel == "Tweet":
		conditions = append(conditions, fmt.Sprintf("n.url = 'http://random.url/%s'", uuid.New().String()[:4]))
		conditions = append(conditions, fmt.Sprintf("m.id_str = '%s'", generateNumericID()))
	case fromLabel == "Tweet" && toLabel == "User":
		conditions = append(conditions, fmt.Sprintf("n.id_str = '%s'", generateNumericID()))
		conditions = append(conditions, fmt.Sprintf("m.url = 'http://random.url/%s'", uuid.New().String()[:4]))
	case fromLabel == "Tweet" && toLabel == "Hashtag":
		conditions = append(conditions, fmt.Sprintf("n.id_str = '%s'", generateNumericID()))
		conditions = append(conditions, fmt.Sprintf("m.name = 'Name_%s'", uuid.New().String()[:4]))
	case fromLabel == "Tweet" && toLabel == "Source":
		conditions = append(conditions, fmt.Sprintf("n.id_str = '%s'", generateNumericID()))
		conditions = append(conditions, fmt.Sprintf("m.name = 'Name_%s'", uuid.New().String()[:4]))
	}

	return fmt.Sprintf(`MATCH (n:%s), (m:%s) WHERE %s AND %s CREATE (n)-[s:%s]->(m) RETURN n,s,m`, fromLabel, toLabel, conditions[0], conditions[1], relationshipType)
}

// generateNumericID generates a random numeric ID
func generateNumericID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Intn(1000000000))
}

func main() {
	// Open file for writing
	file, err := os.Create("generated_queries.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Redirect output to file
	fmt.Println("Generated queries are written into generated_queries.txt")
	os.Stdout = file

	// Store created node IDs
	var nodeIDs []string

	// Generate CREATE queries for nodes
	for i := 0; i < 5; i++ {
		label := NodeLabels[rand.Intn(len(NodeLabels))]
		properties := GenerateRandomProperties(label)
		if label == "Tweet" {
			nodeIDs = append(nodeIDs, properties["id_str"].(string))
		}
		createQuery := GenerateCreateQuery(label, properties)
		fmt.Println(createQuery)
	}

	// Generate relationship creation queries
	for relationshipType, fromToLabels := range RelationshipTypes {
		for fromLabel, toLabels := range fromToLabels {
			for _, toLabel := range toLabels {
				relationshipQuery := GenerateRelationshipQuery(relationshipType, fromLabel, toLabel)
				fmt.Println(relationshipQuery)
			}
		}
	}
}
