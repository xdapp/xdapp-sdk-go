package server

import (
	"../register"
	"net"
	"log"
)

// Client holds info about connection
type Client struct {
	conn   net.Conn
	Server *Server
}

// TCP server
type Server struct {
	address                  	string // Address 如: localhost:9999

	// 调用时重写回调方法
	OnConnectCallback    	func(c *Client)
	OnCloseCallback 		func(c *Client, err error)
	OnReceiveCallback       func(c *Client, message []byte)
	OnReaderCallback   		func(c *Client, message string)
}

// Read client data from channel
func (c *Client) listen() {

	scanner := Spilt(c.conn)

	for scanner.Scan() {
		c.Server.OnReceiveCallback(c, scanner.Bytes())
	}

	if err := scanner.Err(); err != nil {
		c.Server.OnCloseCallback(c, err)
		return
	}
}

// Send text message to client
func (c *Client) Send(message string) error {
	_, err := c.conn.Write([]byte(message))
	return err
}

// Send bytes to client
func (c *Client) SendBytes(b []byte) error {
	_, err := c.conn.Write(b)
	return err
}

func (c *Client) Conn() net.Conn {
	return c.conn
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Called right after server starts listening new client
func (s *Server) OnConnect(callback func(c *Client)) {
	s.OnConnectCallback = callback
}

// Called right after connection closed
func (s *Server) OnClose(callback func(c *Client, err error)) {
	s.OnCloseCallback = callback
}

// Called when Client receives new message
func (s *Server) OnReceive(callback func(c *Client, message []byte)) {
	s.OnReceiveCallback = callback
}

// Start network server
func (s *Server) Listen() {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatal("Error starting TCP server.")
	}
	defer listener.Close()

	for {
		conn, _ := listener.Accept()
		client := &Client{
			conn:   conn,
			Server: s,
		}
		go client.listen()
		s.OnConnectCallback(client)
	}
}

// Creates new tcp server instance
func New(address string) *Server {
	log.Println("Creating tcp server with address", address)

	server := &Server{
		address: address,
	}

	// 初始化
	server.OnConnect(func(c *Client) {})
	server.OnReceive(func(c *Client, message []byte) {})
	server.OnClose(func(c *Client, err error) {})

	return server
}