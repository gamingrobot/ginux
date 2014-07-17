package main

import (
    "encoding/json"
    "github.com/go-martini/martini"
    "github.com/gorilla/websocket"
    "github.com/martini-contrib/sessions"
    "log"
    "net/http"
    "os"
    "fmt"
    "sync"
)

type Config struct {
    Secret string
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

    m.Get("/get", func(session sessions.Session) string {
        v := session.Get("vm_id")
        if v == nil {
            return ""
        }
        return fmt.Sprintf("%d", v)
    })

    m.Get("/create", func(w http.ResponseWriter, r *http.Request, session sessions.Session) string {
        //create vm
        db := GetDB()
        res, e := db.Exec("INSERT INTO `vms_allocated` (`id`) VALUES (NULL);")
        if e != nil {
            return fmt.Sprintf("DB Error: %s", e.Error())
        }
        mid, e := res.LastInsertId()
        if e != nil {
            return fmt.Sprintf("Error: %s", e.Error())
        }

        err := vzcontrol.CreateContainer(mid)
        if err != nil {
            return e.Error()
        }
        session.Set("vm_id", mid)
        return fmt.Sprintf("%d", mid)
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
        err = vzcontrol.StartConsole(vm_id)
        if err != nil {
            ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
            return
        }
        for {
            _, message, err := ws.ReadMessage()
            if err != nil {
                err = vzcontrol.KillConsole(vm_id)
                log.Println(err)
                return
            } else {
                vzcontrol.WriteConsole(vm_id, message)
            }
        }
    })
    log.Println("Game Server started on", config.Address)
    log.Fatal(http.ListenAndServe(config.Address, m))
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
