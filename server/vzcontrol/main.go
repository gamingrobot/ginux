package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kr/pty"
	"io"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"sync"
	"time"
	"unicode/utf8"
)

type Console struct {
	file    *os.File
	command *exec.Cmd
}

type LockingConsoles struct {
	mutex sync.RWMutex
	byId  map[int64]Console
}

var consoles *LockingConsoles

func (c *LockingConsoles) deleteConsole(id int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.byId, id)
}

func (c *LockingConsoles) addConsole(id int64, console Console) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.byId[id] = console
}

type ConsoleChunk struct {
	Id   int64
	Data []byte
}

var readChannel chan ConsoleChunk

func consoleReadLoop(output *os.File, id int64) {
	for {
		b := make([]byte, 1024)
		_, err := output.Read(b)
		if err == io.EOF {
			return
		}
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		out := fixUTF(string(b))
		readChannel <- ConsoleChunk{
			Id:   id,
			Data: []byte(out),
		}
	}
}

func fixUTF(s string) string {
	if !utf8.ValidString(s) {
		v := make([]rune, 0, len(s))
		for i, r := range s {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(s[i:])
				if size == 1 {
					continue
				}
			}
			v = append(v, r)
		}
		s = string(v)
	}
	return s
}

func consoleWriter(r io.Reader) {
	buffer := bufio.NewReader(r)
	for {
		str, err := buffer.ReadString('\n')
		if err != nil {
			continue
		}
		var chunk ConsoleChunk
		json.Unmarshal([]byte(str), &chunk)
		consoles.mutex.RLock()
		consoles.byId[chunk.Id].file.Write(chunk.Data)
		consoles.mutex.RUnlock()
	}
}

func consoleReader(c net.Conn) {
	for chunk := range readChannel {
		str, _ := json.Marshal(chunk)
		output := string(str) + "\n"
		_, err := c.Write([]byte(output))
		if err != nil {
			log.Println("Write: " + err.Error())
		}
	}
}

//console socket
func consoleListen() {
	l, err := net.Listen("unix", "@/tmp/vzconsole.sock")
	if err != nil {
		fmt.Println("listen error", err.Error())
		return
	}

	for {
		fd, err := l.Accept()
		if err != nil {
			fmt.Println("accept error", err.Error())
			return
		}
		go consoleReader(fd)
		go consoleWriter(fd)

	}
}

//rpc socket
func main() {
	log.Println("Started VZControl")
	consoles = &LockingConsoles{
		byId: make(map[int64]Console),
	}
	readChannel = make(chan ConsoleChunk)
	go consoleListen()
	vz := new(VZControl)
	rpc.Register(vz)
	listener, e := net.Listen("unix", "@/tmp/vzcontrol.sock")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	for {
		if conn, err := listener.Accept(); err != nil {
			log.Fatal("accept error: " + err.Error())
		} else {
			log.Printf("new connection established\n")
			go rpc.ServeConn(conn)
		}
	}
}

type VZControl struct{}

func (vz *VZControl) ContainerCreate(cid int64, reply *int64) error {
	err_create := createContainer(cid)
	if err_create != nil {
		return errors.New(fmt.Sprintf("Create Error: %s", err_create.Error()))
	}
	err_mount := setupMount(cid)
	if err_mount != nil {
		return errors.New(fmt.Sprintf("Mount Error: %s", err_mount.Error()))
	}
	err_start := startContainer(cid)
	if err_start != nil {
		return errors.New(fmt.Sprintf("Start Error: %s", err_start.Error()))
	}
	reply = &cid
	return nil
}

func (vz *VZControl) ConsoleStart(cid int64, reply *int64) error {
	err := startConsole(cid)
	if err != nil {
		return errors.New(fmt.Sprintf("Console Start Error: %s", err.Error()))
	}
	reply = &cid
	return nil
}

func (vz *VZControl) ConsoleKill(cid int64, reply *int64) error {
	err := killConsole(cid)
	if err != nil {
		return errors.New(fmt.Sprintf("Console Kill Error: %s", err.Error()))
	}
	reply = &cid
	return nil
}

func (vz *VZControl) NetworkCreate(networkid int64, reply *int64) error {
	err := addBridge(networkid)
	if err != nil {
		return errors.New(fmt.Sprintf("Create Network Error: %s", err.Error()))
	}
	reply = &networkid
	return nil
}

type NetworkAddArgs struct {
    Id, NetworkId int64
}

func (vz *VZControl) NetworkAdd(args *NetworkAddArgs, reply *int64) error {
	cid := args.Id
	networkid := args.NetworkId
	err := addInterface(cid, networkid)
	if err != nil {
		return errors.New(fmt.Sprintf("Interface Add Error: %s", err.Error()))
	}
	err = connectBridge(cid, networkid)
	if err != nil {
		return errors.New(fmt.Sprintf("Bridge Connect Error: %s", err.Error()))
	}
	reply = &cid
	return nil
}

func createContainer(id int64) error {
	command := exec.Command("vzctl", "create", fmt.Sprintf("%d", id), "--config", "ginux")
	err := command.Run() //blocking
	return err
}

func setupMount(id int64) error {
	command := exec.Command("cp", "/etc/vz/conf/ginux.mount", fmt.Sprintf("/etc/vz/conf/%d.mount", id))
	err := command.Run()
	return err
}

func startContainer(id int64) error {
	command := exec.Command("vzctl", "start", fmt.Sprintf("%d", id))
	err := command.Run() //blocking
	return err
}

func addInterface(id int64, networkid int64) error {
	command := exec.Command("addeth.sh", fmt.Sprintf("%d", id),  fmt.Sprintf("%d", networkid))
	err := command.Run()
	return err
}

func addBridge(networkid int64) error {
	command := exec.Command("addbr.sh", fmt.Sprintf("%d", networkid))
	err := command.Run()
	return err
}

func connectBridge(id int64, networkid int64) error {
	command := exec.Command("brctl", fmt.Sprintf("vzbr%d", networkid), fmt.Sprintf("veth%d.%d", id, networkid))
	err := command.Run()
	return err
}

func startConsole(id int64) error {
	consoles.mutex.RLock()
	_, exists := consoles.byId[id]
	consoles.mutex.RUnlock()
	if exists {
		return errors.New(fmt.Sprintf("Console %d already is open", id))
	}
	cmd := exec.Command("vzctl", "console", fmt.Sprintf("%d", id))
	f, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	con := Console{
		file:    f,
		command: cmd,
	}
	consoles.addConsole(id, con)
	go consoleReadLoop(f, id)
	return nil
}

func killConsole(id int64) error {
	consoles.mutex.RLock()
	console, ok := consoles.byId[id]
	consoles.mutex.RUnlock()
	if ok {
		if console.command != nil {
			err_kill := console.command.Process.Kill()
			if err_kill != nil {
				return err_kill
			}
			_, err_wait := console.command.Process.Wait()
			if err_wait != nil {
				return err_wait
			}
		}
	}
	consoles.deleteConsole(id)
	return nil
}
