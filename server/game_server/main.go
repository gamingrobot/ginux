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
	"strconv"
	"sync"
	"time"
)

const CLEAR string = "\033[H\033[2J"
const RESET string = "\033c\033(B\033[0m\033[J\033[?25h"

const MAX_CONSOLE int = 10000

const (
	WSTerm  = 1
	WSClick = 2
)

type Config struct {
	Secret  string
	Address string
}

type LockingWebsockets struct {
	sync.RWMutex
	byId        map[int64]*websocket.Conn
	consoleToId map[int64][]int64
	currentId   int64
}

var consoleBuffers map[int64]string

var websockets *LockingWebsockets

func (c *LockingWebsockets) deleteWebsocket(id int64) {
	c.Lock()
	defer c.Unlock()
	delete(c.byId, id)
}

func (c *LockingWebsockets) addWebsocket(ws *websocket.Conn) int64 {
	c.Lock()
	defer c.Unlock()
	c.currentId += 1
	retid := c.currentId
	c.byId[retid] = ws
	return retid
}

func main() {
	consoleBuffers = make(map[int64]string)
	websockets = &LockingWebsockets{
		byId:        make(map[int64]*websocket.Conn),
		consoleToId: make(map[int64][]int64),
		currentId:   0,
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
		err = vzcontrol.ConsoleStart(int64(startNode.Id))

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
					err = vzcontrol.ConsoleStart(int64(targetNode.Id))
					if err != nil {
						return fmt.Sprintf("Console Start: %d\n%s", targetNode.Id, err.Error())
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
		var currentVm int64 = -1
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
		websocketId := websockets.addWebsocket(ws)
		defer websockets.deleteWebsocket(websocketId)
		ws.WriteMessage(websocket.TextMessage, []byte("Welcome to ginux!\r\n"))
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			} else {
				msgType := message[0]
				msgData := message[1:len(message)]
				switch msgType {
				case WSTerm:
					if currentVm != -1 {
						vzcontrol.ConsoleWrite(currentVm, msgData)
					}
				case WSClick:
					prevVm := currentVm
					tmp, _ := strconv.Atoi(string(msgData))
					currentVm = int64(tmp)
					websockets.Lock()
					if prevVm != -1 {
						for index, wsId := range websockets.consoleToId[prevVm] {
							if wsId == websocketId {
								websockets.consoleToId[prevVm] = append(websockets.consoleToId[prevVm][:index], websockets.consoleToId[prevVm][index+1:]...)
							}
						}
					}
					websockets.consoleToId[currentVm] = append(websockets.consoleToId[currentVm], websocketId)
					websockets.Unlock()
					//ws.WriteMessage(websocket.TextMessage, []byte(CLEAR))
					ws.WriteMessage(websocket.TextMessage, []byte(RESET))
					//ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Selected Container %d\r\n", currentVm)))
					ws.WriteMessage(websocket.TextMessage, []byte(consoleBuffers[currentVm]))
				}
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
		consoleBuffers[chunk.Id] += string(chunk.Data)
		//if len(consoleBuffers[chunk.Id]) > MAX_CONSOLE { 
		//	consoleBuffers[chunk.Id] = consoleBuffers[chunk.Id][len(string(chunk.Data)):]
		//}
		websockets.RLock()
		for _, wsId := range websockets.consoleToId[chunk.Id] {
			if socket, ok := websockets.byId[wsId]; ok {
				socket.WriteMessage(websocket.TextMessage, chunk.Data)
			}
		}
		websockets.RUnlock()
	}
}
