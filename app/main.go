package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"log"
	"net"
	"net/url"
	"time"
)

const (
	azureStorageAccountName = "aisdatansp"
	azureStorageAccessKey   = "n7Cr9caLWJM+tFRSVzcqFmgRRCHKyHYhtm7u3RuBeS88XWp/mAY7hUqP13oTY5K9SOBmGfHu2XIS+AStxQnHxw=="
	containerName           = "test"
)

func main() {
	log.Println("Server is running on port 2001")

	listener, err := net.Listen("tcp", ":2001")
	if err != nil {
		log.Fatal(err)
	}

	// Azure Event Hubs configuration
	// Create an Event Hubs producer client using the connection string and event hub name
	namespaceConnectionString := "Endpoint=sb://aiseventhubs.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=evrLLqNpWg6VlHSH8+eAvha//GtM9mP6b+AEhAQtHrY="
	eventHubName := "ais-data-eventhub"

	defer listener.Close()

	for {
		acceptConnection(listener, namespaceConnectionString, eventHubName)
	}
}

func acceptConnection(listener net.Listener, namespaceConnectionString, eventHubName string) {
	conn, err := listener.Accept()
	if err != nil {
		log.Println(err)
		return
	}

	go handleConnection(conn, namespaceConnectionString, eventHubName)
}

func handleConnection(conn net.Conn, namespaceConnectionString, eventHubName string) {
	defer conn.Close()

	log.Printf("Client connected %s", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		log.Printf("Received message: %s\n", message)
		go sendToBlobStorage(message)
		go processClientMessage(namespaceConnectionString, eventHubName, message)
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading:", err)
	}
}

func sendToBlobStorage(aisMessage string) {
	// use the constants to access the blob storage and add the aisMessage to the blob storage
	credential, err := azblob.NewSharedKeyCredential(azureStorageAccountName, azureStorageAccessKey)
	if err != nil {
		log.Printf("Error creating shared key credential: %v", err)
		return
	}

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", azureStorageAccountName, containerName))
	containerURL := azblob.NewContainerURL(*u, p)

	// Create a unique name for the blob
	blobName := fmt.Sprintf("AISMessage_%s.txt", time.Now().Format("20060102150405"))

	// Get a blob URL reference
	blobURL := containerURL.NewBlockBlobURL(blobName)

	// Convert the AIS message to bytes
	data := []byte(aisMessage)

	// Create a context to control the request execution
	ctx := context.Background()

	// Create the blob
	_, err = azblob.UploadBufferToBlockBlob(ctx, data, blobURL, azblob.UploadToBlockBlobOptions{})
	if err != nil {
		log.Printf("Error uploading to Blob Storage: %v", err)
	}
}

func processClientMessage(namespaceConnectionString, eventHubName, message string) {
	// Process the message received from the client and send it to the Event Hub
	err := triggerEventHub(namespaceConnectionString, eventHubName, message)
	if err != nil {
		log.Printf("Error sending message to Event Hub: %v", err)
	}
}

func triggerEventHub(namespaceConnectionString, eventHubName, aisMessage string) error {
	producerClient, err := azeventhubs.NewProducerClientFromConnectionString(namespaceConnectionString, eventHubName, nil)
	if err != nil {
		return err
	}
	defer producerClient.Close(context.TODO())

	// Create a batch for the AIS message
	events := messageToEvent(aisMessage)

	newBatchOptions := &azeventhubs.EventDataBatchOptions{}
	batch, err := producerClient.NewEventDataBatch(context.TODO(), newBatchOptions)
	if err != nil {
		return err
	}

	for i := 0; i < len(events); i++ {
		err = batch.AddEventData(events[i], nil)
		if err != nil {
			return err
		}
	}

	// Send the batch of events to the Event Hub
	return producerClient.SendEventDataBatch(context.TODO(), batch, nil)
}

func messageToEvent(aisMessage string) []*azeventhubs.EventData {
	return []*azeventhubs.EventData{
		{
			Body: []byte(aisMessage),
		},
	}
}
