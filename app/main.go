package main

import (
	"bufio"
	"context"
	"log"
	"net"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
)

func main() {
	log.Println("Server is running on port 2001")

	listener, err := net.Listen("tcp", ":2001")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	for {
		acceptConnection(listener)
	}
}

func acceptConnection(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		log.Println(err)
		return
	}

	go handleConnection(conn)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("Client connected %s", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		log.Printf("Received message: %s\n", message)
		go processClientMessage(message)
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading:", err)
	}
}

func processClientMessage(message string) {
	// Process the message received from the client and send it to the Event Hub
	err := triggerEventHub(message)
	if err != nil {
		log.Printf("Error sending message to Event Hub: %v", err)
	}
}

func triggerEventHub(aisMessage string) error {
	eventHubNamespace := "aiseventhubs.servicebus.windows.net"
	eventHubName := "ais-data-eventhub"

	defaultAzureCred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	// connectionString := "Endpoint=sb://aiseventhubs.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=evrLLqNpWg6VlHSH8+eAvha//GtM9mP6b+AEhAQtHrY="
	// producerClient, err := azeventhubs.NewProducerClientFromConnectionString(namespaceConnectionString, eventHubName, nil)

	producerClient, err := azeventhubs.NewProducerClient(eventHubNamespace, eventHubName, defaultAzureCred, nil)
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

	if batch.NumEvents() > 0 {
		err := producerClient.SendEventDataBatch(context.Background(), batch, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func messageToEvent(aisMessage string) []*azeventhubs.EventData {
	return []*azeventhubs.EventData{
		{
			Body: []byte(aisMessage),
		},
		{
			Body: []byte(aisMessage),
		},
	}
}
