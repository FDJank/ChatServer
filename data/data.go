package data

import "net"

type Message struct {
	User    string
	Content string
	Addr    string
}

type ClientInfo struct {
	Addr       string
	User       string
	ClientConn net.Conn
}

type ChangeClient struct {
	Addr       string
	IsChange   bool
	User       string
	ClientConn net.Conn
}
