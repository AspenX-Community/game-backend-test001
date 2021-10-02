package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	socketio "github.com/googollee/go-socket.io"
)

type Usuario struct {
	id   string
	nome string
	com  *socketio.Conn
}

// Lista de Usuarios conectados
var clientes map[string]Usuario

func main() {

	server := socketio.NewServer(nil)

	clientes = make(map[string]Usuario)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		//id := uuid.New().String()

		clientes[s.ID()] = Usuario{
			id:  s.ID(),
			com: &s,
		}
		fmt.Println("connected:", s.ID())
		//PrintMemUsage()
		fmt.Println("Clientes: ", len(clientes))
		return nil
	})
    // LATENCIA 
	server.OnEvent("/", "ping", func(s socketio.Conn, msg string) {
		s.Emit("pong")
	})

	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		fmt.Println("notice:", msg)

		for id := range clientes {
			var com = *clientes[id].com
			com.Emit("reply", ""+msg)
		}
		fmt.Println("Clientes: ", len(clientes))
		//PrintMemUsage()
	})

	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		fmt.Println("Clientes: ", len(clientes))
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
		fmt.Println("Clientes: ", len(clientes))
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
func PrintMemUsage() {
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
}
