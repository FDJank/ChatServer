package main

import (
	"ChatServer/data"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

var client_list = make(map[string]data.ClientInfo)

var send_content = make(chan []byte)

// 变更操作client_list的通道
var change_client = make(chan data.ChangeClient)

func read(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		addr := conn.RemoteAddr().String()
		client_info, is_exists := client_list[addr]

		if err != nil {
			//删除客户端列表
			if is_exists {
				go closeClient(conn)
			}
			break
		} else {
			if !is_exists {
				go func() {
					change_client <- data.ChangeClient{
						Addr:       addr,
						IsChange:   true,
						User:       client_info.User,
						ClientConn: conn,
					}
				}()
			}
		}

		content := strings.TrimSpace(string(buf[:n]))
		fmt.Printf("读取到信息：%s\n", content)

		//发送信息到发送通道
		msg, err := json.Marshal(data.Message{
			User:    client_info.User,
			Content: content,
			Addr:    addr,
		})
		if err != nil {
			fmt.Println(msg, " ", err)
			continue
		}

		send_content <- msg
	}
}

func write() {
	for {
		msg := <-send_content
		var message data.Message
		json.Unmarshal(msg, &message)
		for addr, clien_info := range client_list {
			if addr != message.Addr {
				_, err := clien_info.ClientConn.Write(msg)
				if err != nil {
					go func() {
						closeClient(clien_info.ClientConn)
					}()

					continue
				}
			}
		}
	}
}

func closeClient(conn net.Conn) {
	fmt.Printf("'%s'用户关闭连接\n", conn.RemoteAddr().String())
	defer conn.Close()
	change_client <- data.ChangeClient{
		Addr:       conn.RemoteAddr().String(),
		IsChange:   false,
		User:       "",
		ClientConn: conn,
	}
}

func clientHandler(conn net.Conn) {
	msg := data.Message{
		User:    "",
		Content: "请输入名称：",
		Addr:    conn.RemoteAddr().String(),
	}
	reconnect := 3
	for {
		byte_data, err := json.Marshal(msg)
		if err == nil {
			conn.Write(byte_data)
			break
		}
		reconnect--
		if reconnect <= 1 {
			return
		}
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	user_name := strings.TrimSpace(string(buf[:n]))
	fmt.Printf("读取到信息(handler)：%s\n", user_name)
	addr := conn.RemoteAddr().String()
	fmt.Printf("连接的设备：%s\n", addr)
	change_client <- data.ChangeClient{
		Addr:       conn.RemoteAddr().String(),
		IsChange:   true,
		User:       user_name,
		ClientConn: conn,
	}

	go read(conn)
}

func changeClientInfo() {
	for {
		c_client := <-change_client
		_, is_exists := client_list[c_client.Addr]
		if c_client.IsChange {
			client_list[c_client.Addr] = data.ClientInfo{
				Addr:       c_client.Addr,
				User:       c_client.User,
				ClientConn: c_client.ClientConn,
			}
		} else {
			if is_exists {
				delete(client_list, c_client.Addr)
			}
		}
		fmt.Println(client_list)
	}
}

func main() {
	var listen net.Listener
	// 连接失败就重新连接3次
	reconnect := 3
	for {
		var err error
		listen, err = net.Listen("tcp", "127.0.0.1:9000")
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("服务器端启动成功!")
			break
		}

		reconnect--
		if reconnect <= 1 {
			fmt.Println("连接失败!")
			return
		}
	}

	go changeClientInfo()
	go write()

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go clientHandler(conn)
	}

}
