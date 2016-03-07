// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Event Socket client that connects to FreeSWITCH to originate a new call.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"

	"github.com/fiorix/go-eventsocket/eventsocket"
)

type server struct {
	Host     string //Freeswitch hostname
	Port     int    //Freeswitch port
	Password string //Freeswitch password
	Timeout  int    //Freeswitch conneciton timeout in seconds
}

type message struct {
	Host    string //IP of the server that a message is being sent to
	Message string //message needs to be sent
}

func handleServer(fsServer server, eventChannel chan eventsocket.Event, messagesChannel chan message, done chan bool, startedGoroutines *int) {

	//create the full server name as well as destinaion and dialplan if needed at some point
	fullServer := fsServer.Host + ":" + strconv.Itoa(fsServer.Port)
	// dest := "sofia/mydomain.com/1000@" + fsServer.Host
	// dialplan := "&socket(" + fsServer.Host + ":9090 async full)"

	//connects to the server
	c, err := eventsocket.Dial(fullServer, fsServer.Password)
	if err != nil {
		log.Fatal(err)
	}

	//at the end it closes the connection to the server
	defer c.Close()

	//set which messages
	c.Send("events json ALL")

	//sends the event
	// c.Send(fmt.Sprintf("bgapi originate %s", dest))
	for {
		ev, err := c.ReadEvent()
		if err != nil {
			log.Fatal(err)
		}

		//checks what to do next with the message
		select {
		case eventChannel <- *ev:
			//sends the event to the channel

		case message := <-messagesChannel:
			if message.Host == fsServer.Host {
				fmt.Println("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\nmessage host:\t" + message.Host + "\nserver host:\t" + fsServer.Host + "\nmessage sent:\t" + message.Message + "\nXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
			}
		case <-done:
			*startedGoroutines--
			fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\nCLOSED server " + fsServer.Host + " goroutine, " + strconv.Itoa(*startedGoroutines) + " goroutines left to close\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
			return
		}

		//prints out the event
		fmt.Println("\nNew event")
		ev.PrettyPrint()

		//checks if a hangup happened and returns the function
		if ev.Get("Answer-State") == "hangup" {
			*startedGoroutines--
			// break
			return
		}
	}
}

func handleMessage(eventChannel chan eventsocket.Event, messagesChannel chan message, done chan bool, startedGoroutines *int) {

	var event eventsocket.Event

	for {

		if *startedGoroutines == 1 {
			*startedGoroutines--
			fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\nALL other goroutines closed, closing messages goroutine\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
			return
		}

		select {

		case event = <-eventChannel:
			fmt.Print("----------------------------------------------------------\n" + "UUID:\t\t" + event.Header["Core-Uuid"] + "\nIP:\t\t" + event.Header["Freeswitch-Ipv4"] +
				"\nFunction:\t" + event.Header["Event-Calling-Function"] + "\nTime:\t\t" + event.Header["Event-Date-Timestamp"] + "\nSequence#:\t" + event.Header["Event-Sequence"])

			if event.Body != "" {
				//create a new function that deals with the body of the event
				fmt.Println("\nBody exists and it needs to be processed\n----------------------------------------------------------")
			} else {
				fmt.Println("\nBody is empty\n----------------------------------------------------------")
			}

			if event.Header["Event-Name"] == "HEARTBEAT" {
				messagesChannel <- message{event.Header["Freeswitch-Ipv4"], "bgapi originate test"}
			}

		case <-done:
			*startedGoroutines--
			fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\nCLOSED messages goroutine, " + strconv.Itoa(*startedGoroutines) + " goroutines left to close\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
			return
		}
	}
}

func startServers(fsServers []server, eventChannel chan eventsocket.Event, messagesChannel chan message, done chan bool, startedGoroutines *int) {

	//channel used for sending messages that are received
	// eventChannel = make(chan eventsocket.Event)

	for _, v := range fsServers {

		*startedGoroutines++
		go handleServer(v, eventChannel, messagesChannel, done, startedGoroutines)
	}

	*startedGoroutines++
	go handleMessage(eventChannel, messagesChannel, done, startedGoroutines)
}

func main() {

	//define servers, this can be passed in as json or parameters through an API or read in from a database
	//for now only 2 servers are defined
	var fsServers []server
	fsServers = append(fsServers, server{"208.76.55.72", 8021, "ClueCon", 10})
	fsServers = append(fsServers, server{"208.76.55.73", 8021, "ClueCon", 10})

	//sets the maximum number of operating system threads that can execute user-level Go code simultaneously
	runtime.GOMAXPROCS(runtime.NumCPU())

	//channel used for sending messages that are received
	eventChannel := make(chan eventsocket.Event)
	defer close(eventChannel)

	//channel used for sending messages back to the server
	messagesChannel := make(chan message)
	defer close(messagesChannel)

	//channel used to stop goroutines
	done := make(chan bool, len(fsServers)+1)
	startedGoroutines := 0

	// done := make(chan struct{}, 1)
	defer close(done)

	go startServers(fsServers, eventChannel, messagesChannel, done, &startedGoroutines)

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		if startedGoroutines > 0 {
			fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\nGOROUTINES ALREADY STARTED, STOP FIRST\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		} else {
			startedGoroutines = 0
			fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\nSTARTING ALL GOROUTINES\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
			go startServers(fsServers, eventChannel, messagesChannel, done, &startedGoroutines)
		}
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		if startedGoroutines == 0 {
			fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\nALL GOROUTINES ALREADY STOPPED, START FIRST\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		} else {
			fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\nCLOSING ALL " + strconv.Itoa(startedGoroutines) + " GOROUTINES\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
			for i := 0; i < startedGoroutines; i++ {
				done <- true
			}
		}
	})

	http.HandleFunc("/exit", func(w http.ResponseWriter, r *http.Request) {
		if startedGoroutines > 0 {
			fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\nFIRST STOP GOROUTINES\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		} else {
			os.Exit(0)
		}
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
