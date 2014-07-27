package main

import (
	"encoding/json"
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

	m.Get("/reset", func(w http.ResponseWriter, r *http.Request, session sessions.Session) string {
		return "No"
	})

	generating := false
	gr := NewGraph()
	m.Get("/gen", func(w http.ResponseWriter, r *http.Request, session sessions.Session) string {
		if generating {
			return "Already generating"
		}
		generating = true
		maxNodes := 20
		maxEdges := 10
		startNodeId := 101

		startNode := Node{Id: NodeId(startNodeId)}
		gr.AddNode(startNode)
        vzcontrol.ContainerCreate(int64(startNode.Id))

		nodes := make([]Node, 0)
		nodes = append(nodes, startNode)

		steps := 1
		for len(nodes) != 0 && steps < maxNodes {
			node, nodes := nodes[len(nodes)-1], nodes[:len(nodes)-1]

			numEdges := random(1, maxEdges)
			for i := 1; i <= numEdges; i++ {
				targetNode := Node{Id: NodeId(i*steps + startNodeId)}
				gr.AddNode(targetNode)
                vzcontrol.ContainerCreate(int64(targetNode.Id))
				nodes = append(nodes, targetNode)
                edgeid := int64(i*steps)
				gr.AddEdge(Edge{Id: EdgeId(edgeid), Head: node.Id, Tail: targetNode.Id})
                vzcontrol.NetworkCreate(edgeid)
                vzcontrol.NetworkAdd(int64(node.Id), edgeid)
                vzcontrol.NetworkAdd(int64(targetNode.Id), edgeid)

			}
			steps += 1
		}

		/*err := vzcontrol.ContainerCreate(101)
		  if err != nil {
		      return err.Error()
		  }
		  err = vzcontrol.ContainerCreate(102)
		  if err != nil {
		      return err.Error()
		  }
		  err = vzcontrol.NetworkCreate(0)
		  if err != nil {
		      return err.Error()
		  }
		  err = vzcontrol.NetworkAdd(101, 0)
		  if err != nil {
		      return err.Error()
		  }
		  err = vzcontrol.NetworkAdd(102, 0)
		  if err != nil {
		      return err.Error()
		  }*/
		return gr.String()
	})

	m.Get("/ws", func(w http.ResponseWriter, r *http.Request, session sessions.Session) {
		return //FOR NOW JUST IGNORE WEBSOCKETS
		/*ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
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
		  //get vm numbert
		  vm_cookie := session.Get("vm_id")
		  var vm_id int64
		  if vm_cookie == nil {
		      ws.WriteMessage(websocket.TextMessage, []byte("Create a node at ginux.gamingrobot.net/create\r\n"))
		      return
		  } else {
		      vm_id = vm_cookie.(int64)
		  }

		  _, exists := websockets.byId[vm_id];
		  if !exists{
		      websockets.addWebsocket(vm_id, ws)
		      defer websockets.deleteWebsocket(vm_id)
		  }
		  //spawn console
		  log.Println(vm_id)
		  err = vzcontrol.ConsoleStart(vm_id)
		  if err != nil {
		      ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		      return
		  }
		  for {
		      _, message, err := ws.ReadMessage()
		      if err != nil {
		          err = vzcontrol.ConsoleKill(vm_id)
		          log.Println(err)
		          return
		      } else {
		          vzcontrol.ConsoleWrite(vm_id, message)
		      }
		  }*/
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
