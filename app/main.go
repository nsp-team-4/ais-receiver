package main

import (
	"bufio"
	"context"
	"log"
	"net"

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
	triggerEventHub("!AIVDM,1,1,,B,33aEP2hP00PBLRFMfCp;OOw<R>`<,0*49")

	defer listener.Close()

	for {
		acceptConnection(listener)
	}
}

func triggerEventHub(aisMessage string) {
	namespaceConnectionString := "Endpoint=sb://nsp-ais.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=rz7Wmc79oc07VfWLlsLy4SR8iBbS7i7JA+AEhJBUnB0="
	eventHubName := "nsp-ais"

	producerClient, err := azeventhubs.NewProducerClientFromConnectionString(namespaceConnectionString, eventHubName, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer producerClient.Close(context.TODO())

	// create sample AIS message
	events := messageToEvent(aisMessage)

	// create a batch object and add sample events to the batch
	newBatchOptions := &azeventhubs.EventDataBatchOptions{}

	batch, err := producerClient.NewEventDataBatch(context.TODO(), newBatchOptions)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(events); i++ {
		err = batch.AddEventData(events[i], nil)
		if err != nil {
			log.Fatal(err)
		}
	}

	// send the batch of events to the event hub
	producerClient.SendEventDataBatch(context.TODO(), batch, nil)

}

func messageToEvent(aisMessage string) []*azeventhubs.EventData {
	return []*azeventhubs.EventData{
		{
			Body: []byte(aisMessage),
		},
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
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading:", err)
	}
}
