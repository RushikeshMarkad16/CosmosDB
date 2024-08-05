package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

type Config struct {
	COSMOS_DB_ENDPOINT string
	COSMOS_DB_KEY      string
}

const (
	databaseName     = "sample-database"
	containerName    = "sample-container"
	partitionKeyPath = "/partitionKey"
)

func main() {

	// Create a context
	ctx := context.Background()
	config := Config{
		COSMOS_DB_ENDPOINT: "https://localhost:8081",
		COSMOS_DB_KEY:      "C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw==",
	}

	// Create a credential using the key
	cred, err := azcosmos.NewKeyCredential(config.COSMOS_DB_KEY)
	if err != nil {
		log.Printf("failed to create credential: %v", err)
		return
	}

	// Create a Cosmos client
	client, err := azcosmos.NewClientWithKey(config.COSMOS_DB_ENDPOINT, cred, nil)
	if err != nil {
		log.Printf("failed to create client: %v", err)
		return
	}

	// Create a database
	_, err = client.CreateDatabase(ctx, azcosmos.DatabaseProperties{ID: databaseName}, nil)
	if err != nil {
		if strings.Contains(err.Error(), "ERROR CODE: 409 Conflict") {
			fmt.Println("Database already exists")
		} else {
			log.Printf("failed to create database: %v", err)
		}
	}
	fmt.Println("Database created")

	// Get the database client
	databaseClient, err := client.NewDatabase(databaseName)
	if err != nil {
		log.Printf("failed to create database: %v", err)
		return
	}

	// Create a container
	containerProperties := azcosmos.ContainerProperties{
		ID: containerName,
		PartitionKeyDefinition: azcosmos.PartitionKeyDefinition{
			Paths: []string{partitionKeyPath},
		},
	}
	throughput := azcosmos.NewManualThroughputProperties(400)
	_, err = databaseClient.CreateContainer(ctx, containerProperties, &azcosmos.CreateContainerOptions{ThroughputProperties: &throughput})
	if err != nil {
		if strings.Contains(err.Error(), "ERROR CODE: 409 Conflict") {
			fmt.Println("Container already exists")
		} else {
			log.Printf("failed to create container: %v", err)
		}
	}
	fmt.Println("Container created")

	// Get the container client
	containerClient, err := databaseClient.NewContainer(containerName)
	if err != nil {
		log.Printf("failed to create container: %v", err)
		return
	}

	// Define a sample item to insert
	item := map[string]interface{}{
		"id":           "1",
		"partitionKey": "sample-partition",
		"name":         "sample-item",
	}

	// Serialize the item to JSON
	itemJSON, err := json.Marshal(item)
	if err != nil {
		log.Printf("failed to marshal item: %v", err)
		return
	}

	// Insert the item
	pk := azcosmos.NewPartitionKeyString("sample-partition")
	_, err = containerClient.UpsertItem(ctx, pk, itemJSON, nil)
	if err != nil {
		log.Printf("failed to upsert item: %v", err)
		return
	}
	fmt.Println("Item upserted")

	// Read the item
	readResp, err := containerClient.ReadItem(ctx, pk, "1", nil)
	if err != nil {
		log.Printf("failed to read item: %v", err)
		return
	}
	fmt.Printf("Item read: %v\n", string(readResp.Value))

	// Update the item
	item["name"] = "updated-item"
	itemJSON, err = json.Marshal(item)
	if err != nil {
		log.Printf("failed to marshal item: %v", err)
		return
	}
	_, err = containerClient.UpsertItem(ctx, pk, itemJSON, nil)
	if err != nil {
		log.Printf("failed to update item: %v", err)
		return
	}
	fmt.Println("Item updated")

	// Delete the item
	_, err = containerClient.DeleteItem(ctx, pk, "1", nil)
	if err != nil {
		log.Printf("failed to delete item: %v", err)
		return
	}
	fmt.Println("Item deleted")
}
