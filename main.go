package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"time"
)

// --- Communication Protocol ---

// PingArgs represents the request payload.
type PingArgs struct {
	ID int
}

// PingReply represents the response payload.
type PingReply struct {
	Message string
}

// Node represents the state of a single server in the cluster.
type Node struct {
	ID int
}

// Ping is a Remote Procedure Call (RPC) method.
// Rules for Go RPC:
// 1. The method's type is exported (starts with upper case).
// 2. The method has two arguments, both exported.
// 3. The second argument is a pointer.
// 4. The method has return type error.
func (n *Node) Ping(args *PingArgs, reply *PingReply) error {
	fmt.Printf("[Node %d] Received Ping from Node %d\n", n.ID, args.ID)
	reply.Message = fmt.Sprintf("Hello %d, I am Node %d and I am alive!", args.ID, n.ID)
	return nil
}

// --- Bootstrap Logic ---

func main() {
	// 1. Parse Command Line Arguments
	id := flag.Int("id", 0, "Unique ID for the node")
	port := flag.Int("port", 8000, "Port to listen on")
	peer := flag.String("peer", "", "Target peer port to ping (e.g. 8001)")
	flag.Parse()

	// 2. Set up the RPC Server
	node := &Node{ID: *id}
	rpc.Register(node)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(err) // In a real system, we'd handle this more gracefully
	}
	fmt.Printf("Node %d is listening on port %d...\n", *id, *port)

	// Launch the server in a separate Goroutine
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go rpc.ServeConn(conn) // Handle connection in its own thread
		}
	}()

	// 3. Client Logic: Try to connect to a peer if provided
	time.Sleep(2 * time.Second) // Give the other node time to start
	if *peer != "" {
		client, err := rpc.Dial("tcp", "localhost:"+*peer)
		if err == nil {
			args := &PingArgs{ID: *id}
			var reply PingReply
			err = client.Call("Node.Ping", args, &reply)
			if err == nil {
				fmt.Printf(">>> Response from Peer: %s\n", reply.Message)
			}
			client.Close()
		} else {
			fmt.Printf("Could not connect to peer: %v\n", err)
		}
	}

	// Keep the process running forever
	select {}
}
