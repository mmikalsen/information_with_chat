package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"time"

	"github.com/gorilla/websocket"
)

var (
	clients            = make(map[*websocket.Conn]bool) // connected clients
	broadcast          = make(chan Message, 100)        // broadcast channel
	lock               = new(sync.Mutex)
	exPath             string
	informationCounter = 0
	broadcastmsg       = new(Broadcast)
)

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//Message defines our message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
	IP       string `json:"ip"`
}

// Broadcast defines the broadcast message
type Broadcast struct {
	Message string `json:"message"`
	Title   string `json:"title"`
	Content string `json:"content, omitempty"`
}

func main() {
	f, err := os.OpenFile("message.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath = filepath.Dir(ex)

	if err := currentBroadcast(); err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	log.SetOutput(f)

	// Create a simple file server
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true
	log.Println("new client")

	err = ws.WriteJSON(broadcastmsg)
	if err != nil {
		log.Printf("error: %v", err)
		lock.Lock()
		delete(clients, ws)
		lock.Unlock()
	}

	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		log.Println(msg.Username, " - ", msg.Email, " - ", msg.Message, " - ", r.RemoteAddr)
		broadcast <- msg
	}
}

func handleMessages() {

	go func() {
		for {
			// Grab the next message from the broadcast channel
			msg := <-broadcast

			if msg.Message != "" {
				// Send it out to every client that is currently connected
				for client := range clients {
					go func(message Message, client *websocket.Conn) {
						err := client.WriteJSON(msg)
						if err != nil {
							log.Printf("error: %v", err)
							client.Close()
							lock.Lock()
							delete(clients, client)
							lock.Unlock()
						}
					}(msg, client)
				}
			}
		}
	}()

	for {

		for client := range clients {
			go func(broadcast *Broadcast, client *websocket.Conn) {
				err := client.WriteJSON(broadcast)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					lock.Lock()
					delete(clients, client)
					lock.Unlock()
				}
			}(broadcastmsg, client)
		}
		for {
			informationCounter++
			if err := currentBroadcast(); err != nil {
				log.Println(err)
				continue
			}
			break
		}
		time.Sleep(time.Second * 25)
	}
}

func currentBroadcast() error {
	files, err := ioutil.ReadDir(exPath + "/broadcast_msg")
	if err != nil {
		return err
	}

	raw, err := ioutil.ReadFile(exPath + "/broadcast_msg/" + files[informationCounter%len(files)].Name())
	if err != nil {
		return err
	}

	if err := json.Unmarshal(raw, broadcastmsg); err != nil {
		return err
	}

	return nil
}
