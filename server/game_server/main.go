package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	"github.com/martini-contrib/sessions"
	"github.com/martini-contrib/gzip"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	"io/ioutil"
	//"runtime/pprof"
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

var consoleBuffers map[int64]*bytes.Buffer

var websockets *LockingWebsockets

var generatingGraph bool
var generatingError string

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
	//pprof stuff
	/*f, err := os.Create("pprof.out")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	go func() {
		time.Sleep(time.Minute * 10)
		pprof.StopCPUProfile()
		log.Fatal("Done")
	}()*/
	generatingGraph = false
	generatingError = ""
	consoleBuffers = make(map[int64]*bytes.Buffer)
	websockets = &LockingWebsockets{
		byId:        make(map[int64]*websocket.Conn),
		consoleToId: make(map[int64][]int64),
		currentId:   0,
	}
	consoleReadChannel = make(chan *ConsoleChunk)
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
	m.Use(gzip.All())

	gr := NewGraph()
	if _, err := os.Stat("graph.json"); err == nil { //file exists
		file, _ := ioutil.ReadFile("graph.json")
    	json.Unmarshal(file, &gr)
    } else {
		go generateGraph(100, vzcontrol, gr)
	}
	m.Get("/reset/:secret", func(w http.ResponseWriter, r *http.Request, params martini.Params, session sessions.Session) string {
		if params["secret"] != config.Secret {
			return ""
		}
		generatingGraph = false
		err := vzcontrol.Reset()
		if err != nil {
			return err.Error()
		}
		gr = NewGraph()
		os.Remove("graph.json")
		return "Done"
	})

	m.Get("/gen", func(w http.ResponseWriter, r *http.Request, session sessions.Session) string {
		if generatingError != "" {
			return fmt.Sprintf("Generation Error: ", generatingError)
		}
		if generatingGraph {
			return fmt.Sprintf("Generating: ", len(gr.Nodes))
		}
		return "Done"
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
		ws.WriteMessage(websocket.TextMessage, []byte("Welcome to ginux!\r\nClick on a node to get started.\r\n"))
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
					ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Selected Container %d\r\n", currentVm)))
					ws.WriteMessage(websocket.TextMessage, consoleBuffers[currentVm].Bytes())
				}
			}
		}
	})
	log.Println("Game Server started on", config.Address)
	log.Fatal(http.ListenAndServe(config.Address, m))
}

func generateGraph(nodes int, vzcontrol *VZControl, gr *Graph){
		if generatingGraph {
			return
		}
		generatingGraph = true
		//maxNodes := 100
		//maxEdges := 5
		startNodeId := 101

		counter := 0
		for counter < nodes {
			if generatingGraph == false {
				generatingError = "Generation Aborted"
				return
			}
			node := Node{Id: NodeId(startNodeId + counter)}
			gr.AddNode(node)
			err := vzcontrol.ContainerCreate(int64(node.Id))
			if err != nil {
				generatingError = fmt.Sprintf("Create Fail: %d\n%s", node.Id, err.Error())
			}
			err = vzcontrol.ConsoleStart(int64(node.Id))
			if err != nil {
				generatingError = fmt.Sprintf("Start Fail: %d\n%s", node.Id, err.Error())
			}
			counter++
		}
		generatingGraph = false

		/*startNode := Node{Id: NodeId(startNodeId)}
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
		}*/

}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func consoleDispatch() {
	for chunk := range consoleReadChannel {
		if _, ok := consoleBuffers[chunk.Id]; !ok {
			consoleBuffers[chunk.Id] = &bytes.Buffer{}
		}
		consoleBuffers[chunk.Id].Write(chunk.Data)
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
