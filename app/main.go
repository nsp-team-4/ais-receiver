package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
)

func main() {
	log.Println("Server is running on port 2001")

	listener, err := net.Listen("tcp", ":2001")
	if err != nil {
		log.Fatal(err)
	}

	// Azure Event Hubs configuration
	// Create an Event Hubs producer client using the connection string and event hub name
	namespaceConnectionString := "Endpoint=sb://nsp-ais.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=rz7Wmc79oc07VfWLlsLy4SR8iBbS7i7JA+AEhJBUnB0="
	eventHubName := "test123"

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
		go processClientMessage(namespaceConnectionString, eventHubName, message)
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading:", err)
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
