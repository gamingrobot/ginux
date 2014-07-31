package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	"github.com/martini-contrib/sessions"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

const CLEAR string = "\\33[H\\33[2J"

type Config struct {
	Secret  string
	Address string
}

type LockingWebsockets struct {
	mutex sync.RWMutex
	byId  map[int64]*websocket.Conn
}

var websockets *LockingWebsockets

func (c *LockingWebsockets) deleteWebsocket(id int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.byId, id)
}

func (c *LockingWebsockets) addWebsocket(id int64, ws *websocket.Conn) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.byId[id] = ws
}

func main() {
	websockets = &LockingWebsockets{
		byId: make(map[int64]*websocket.Conn),
	}
	consoleReadChannel = make(chan ConsoleChunk)
	go consoleDispatch()
	vzcontrol := ConnectVZControl()
	defer vzcontrol.Close()

	file, _ := os.Open("game_server.cfg")
	decoder := json.NewDecoder(file)
	config := Config{}
	decoder.Decode(&config)

	m := martini.Classic()
	store := sessions.NewCookieStore([]byte(config.Secret))
	m.Use(sessions.Sessions("session", store))

	generating := false
	gr := NewGraph()
	m.Get("/reset/:secret", func(w http.ResponseWriter, r *http.Request, params martini.Params, session sessions.Session) string {
		if params["secret"] != config.Secret {
			return ""
		}
		err := vzcontrol.Reset()
		if err != nil {
			return err.Error()
		}
		generating = false
		gr = NewGraph()
		return "Done"
	})

	m.Get("/gen", func(w http.ResponseWriter, r *http.Request, session sessions.Session) string {
		if generating {
			return "Already generating"
		}
		generating = true
		maxNodes := 5
		maxEdges := 5
		startNodeId := 100

		startNode := Node{Id: NodeId(startNodeId)}
		gr.AddNode(startNode)
		err := vzcontrol.ContainerCreate(int64(startNode.Id))

		nodes := make([]Node, 0)
		nodes = append(nodes, startNode)

		steps := 1
		for len(nodes) != 0 && steps < maxNodes {
			node, nodes := nodes[len(nodes)-1], nodes[:len(nodes)-1]

			numEdges := random(1, maxEdges)
			for i := 1; i <= numEdges; i++ {
				targetNode := Node{Id: NodeId(i*steps + startNodeId)}
				if gr.AddNode(targetNode) {
					err = vzcontrol.ContainerCreate(int64(targetNode.Id))
					if err != nil {
						return fmt.Sprintf("Container Create: %d, %d, %d\n%s", targetNode.Id, i*steps, numEdges, err.Error())
					}
					nodes = append(nodes, targetNode)
					edgeid := int64(i * steps)
					if gr.AddEdge(Edge{Id: EdgeId(edgeid), Head: node.Id, Tail: targetNode.Id}) {
						err = vzcontrol.NetworkCreate(edgeid)
						if err != nil {
							return fmt.Sprintf("Network Create: %d\n%s", edgeid, err.Error())
						}
						err = vzcontrol.NetworkAdd(int64(node.Id), edgeid)
						if err != nil {
							return fmt.Sprintf("Network Add Node: %d, %d\n%s", node.Id, edgeid, err.Error())
						}
						err = vzcontrol.NetworkAdd(int64(targetNode.Id), edgeid)
						if err != nil {
							return fmt.Sprintf("Network Add Target: %d, %d\n%s", targetNode.Id, edgeid, err.Error())
						}
					}
				}

			}
			steps += 1
		}
		return gr.String()
	})

	m.Get("/graph", func(w http.ResponseWriter, r *http.Request, session sessions.Session) string {
		output, err := json.Marshal(gr)
		if err != nil {
			return err.Error()
		}
		return string(output)
	})

	m.Get("/ws", func(w http.ResponseWriter, r *http.Request, session sessions.Session) {
		ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			log.Println(err)
			return
		} else if err != nil {
			log.Println(err)
			return
		}
		defer ws.Close()
		ws.WriteMessage(websocket.TextMessage, []byte("Welcome to ginux!\r\n"))
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			} else {
				ws.WriteMessage(websocket.TextMessage, []byte(message))
			}
		}
	})
	log.Println("Game Server started on", config.Address)
	log.Fatal(http.ListenAndServe(config.Address, m))
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func consoleDispatch() {
	for chunk := range consoleReadChannel {
		websockets.mutex.RLock()
		if socket, ok := websockets.byId[chunk.Id]; ok {
			socket.WriteMessage(websocket.TextMessage, chunk.Data)
		}
		websockets.mutex.RUnlock()
	}
}
