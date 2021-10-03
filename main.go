package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	socketio "github.com/googollee/go-socket.io"
)

type Usuario struct {
	Id       string `json:"id"`
	Nome     string `json:"nome"`
	Position string `json:"position"`
	com      *socketio.Conn
}

// Lista de Usuarios conectados
var clientes map[string]Usuario

func main() {

	server := socketio.NewServer(nil)

	clientes = make(map[string]Usuario)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")

		client := Usuario{
			Id:   s.ID(),
			Nome: "BaconVegano" + string(s.ID()),
			com:  &s,
		}
		data, _ := json.Marshal(client)
		for id := range clientes {
			var com = *clientes[id].com
			com.Emit("new-player", string(data))
			dataUser, _ := json.Marshal(clientes[id])
			s.Emit("new-player", string(dataUser))
		}
		clientes[s.ID()] = client

		fmt.Println("connected:", s.ID())
		//PrintMemUsage()
		fmt.Println("Clientes: ", len(clientes))
		return nil
	})
	// LATENCIA
	server.OnEvent("/", "ping", func(s socketio.Conn, msg string) {
		s.Emit("pong")
	})
	server.OnEvent("/", "move-player", func(s socketio.Conn, msg string) {
		client := clientes[s.ID()]
		client.Position = msg
		data, _ := json.Marshal(client)
		for id := range clientes {
			if id == s.ID() {
				continue
			}
			var com = *clientes[id].com
			com.Emit("move-player", string(data))
		}
	})

	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		fmt.Println("notice:", msg)

		for id := range clientes {
			var com = *clientes[id].com
			com.Emit("reply", ""+msg)
		}
	})

	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		return "enviado... " + msg
	})

	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		delete(clientes, s.ID())
		fmt.Println("closed", reason)
	})

	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./web")))
	log.Println("Serving at localhost:8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
/*func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
	fmt.Printf("\tMallocs = %v\n", bToMb(m.Mallocs))
	fmt.Printf("\tFrees = %v\n", bToMb(m.Frees))

}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}*/
