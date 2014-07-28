//Opens two unix sockets served by vzcontrol process running as root
package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/rpc"
)

type VZControl struct {
	rpcClient         *rpc.Client
	consoleConnection net.Conn
}

var consoleReadChannel chan ConsoleChunk

func ConnectVZControl() *VZControl {
	console := connectConsole()
	client := connectControl()
	control := &VZControl{
		rpcClient:         client,
		consoleConnection: console,
	}
	return control
}

func connectConsole() net.Conn {
	c, err := net.Dial("unix", "@/tmp/vzconsole.sock")
	if err != nil {
		panic(err.Error())
	}
	//defer c.Close()
	go consoleReader(c)
	return c
}

type ConsoleChunk struct {
	Id   int64
	Data []byte
}

func consoleReader(r io.Reader) {
	buffer := bufio.NewReader(r)
	for {
		str, err := buffer.ReadString('\n')
		if err != nil {
			continue
		}
		var chunk ConsoleChunk
		json.Unmarshal([]byte(str), &chunk)
		consoleReadChannel <- chunk
	}
}

func connectControl() *rpc.Client {
	client, err := rpc.Dial("unix", "@/tmp/vzcontrol.sock")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	return client
}

func (vz *VZControl) ContainerCreate(id int64) error {
	var reply int64
	err := vz.rpcClient.Call("VZControl.ContainerCreate", id, &reply)
	return err
}

func (vz *VZControl) ConsoleStart(id int64) error {
	var reply int64
	err := vz.rpcClient.Call("VZControl.ConsoleStart", id, &reply)
	return err
}

func (vz *VZControl) ConsoleKill(id int64) error {
	var reply int64
	err := vz.rpcClient.Call("VZControl.ConsoleKill", id, &reply)
	return err
}

func (vz *VZControl) ConsoleWrite(id int64, data []byte) error {
	chunk := ConsoleChunk{
		Id:   id,
		Data: data,
	}
	str, _ := json.Marshal(chunk)
	output := string(str) + "\n"
	vz.consoleConnection.Write([]byte(output))
	return nil
}

func (vz *VZControl) NetworkCreate(networkid int64) error {
	var reply int64
	err := vz.rpcClient.Call("VZControl.NetworkCreate", networkid, &reply)
	return err
}

func (vz *VZControl) NetworkAdd(id int64, networkid int64) error {
	type NetworkAddArgs struct {
		Id, NetworkId int64
	}
	var reply int64
	err := vz.rpcClient.Call("VZControl.NetworkAdd", &NetworkAddArgs{id, networkid}, &reply)
	return err
}

func (vz *VZControl) Reset() error {
	var reply int64
	err := vz.rpcClient.Call("VZControl.Reset", nil, &reply)
	return err
}


func (vz *VZControl) Close() {
	vz.rpcClient.Close()
}
