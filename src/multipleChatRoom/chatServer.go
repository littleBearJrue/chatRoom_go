package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// 定义聊天室基本数据
type chatRoom struct {
	id int32
	name string

}

// 定义client存储用户的基本数据
type client struct {
	chatChan chan string
	nickName string
	address string
	isOnline bool
}

// 定义一个map表存储在线的所有玩家信息 key：nickName, client:client实体
var onlineClients = make(map[string] client)

func main() {

	fmt.Println("Starting the chat server ...")
	// 启动服务
	startSocket()
}

func startSocket() {
	listener, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		fmt.Println("net listener is error, error is: ", err)
		os.Exit(1)  // 强制退出server端
	}
	// 在此方法调用结束后执行关闭网络连接
	defer listener.Close()

	// 运行在主协程， 阻塞监听客户端的用户连接请求
	for {
		conn, connErr := listener.Accept()
		if connErr != nil {
			fmt.Println("conn error, error is: ", connErr)
			continue
		}

		//// 转发信息到客户端
		//go sendMsgToClient(conn)

		go doServerHandle(conn)

	}
}

//func sendMsgToClient(cli client, conn net.Conn){
//	// 循环遍历每个玩家，将玩家的储存的通道数据一个个通过conn.write转给客户端
//
//	for {
//		for _, clt := range onlineClients {
//			for msgInfo := range clt.chatChan {
//				fmt.Println("write -----> ", clt.nickName, msgInfo )
//				_, err := conn.Write([]byte(msgInfo + "\n"))
//				if err != nil {
//					fmt.Println("conn write is error, error is: ", err)
//				}
//			}
//		}
//	}
//
//}

func sendMsgToClient(clt client, conn net.Conn){
	// 循环遍历每个玩家，将玩家的储存的通道数据一个个通过conn.write转给客户端

	for {
		for msgInfo := range clt.chatChan {
			fmt.Println("write -----> ", clt.nickName, msgInfo )
			_, err := conn.Write([]byte(msgInfo + "\n"))
			if err != nil {
				fmt.Println("conn write is error, error is: ", err)
			}
		}
	}

}

func doServerHandle(conn net.Conn) {
	// 结束后关闭连接
	defer conn.Close()

	// 获取玩家登陆地址
	clientAddr := conn.RemoteAddr().String()
	fmt.Println( clientAddr +  "连接成功")


	for {
		buf := make([]byte, 512)

		len, err := conn.Read(buf)
		if len == 0 || err != nil {
			fmt.Println("Error reading", err)
			return //终止程序
		}
		// 解析客户端发送过来的数据
		msg_str := strings.Split(string(buf[0:len]), "|")  //将从客户端收到的字节流分段保存到msg_str这个数组中
		// msg_str[0] 存放的数数据类型，包括：“online”,“offline”，“say”


		fmt.Println("from client msg: ", msg_str)

		switch msg_str[0] {
		// 玩家登陆上线
		case "online":

			clt := client{make(chan string), msg_str[1], clientAddr, true}
			onlineClients[msg_str[1]] = clt

			go sendMsgToClient(clt, conn)

			fmt.Printf("玩家[%s]上线！", msg_str[1])
			for nickStr, clt := range onlineClients {
				if nickStr != msg_str[1] {
					toMsgChanStr := "玩家" + "[" + msg_str[1] + "]" + "已上线，你们可以欢快的聊天了"
					clt.chatChan <- toMsgChanStr   // 将上线信息传入每个非自己玩家的聊天通道中
				}
			}
		// 玩家的聊天内容，转发给客户端
		case "say":
			fmt.Println("onlineClients_say ==> ", onlineClients)
			for nickStr, clt := range onlineClients {
				if nickStr != msg_str[1] {
					toMsgChanStr := "[" + msg_str[1] + "]： " + msg_str[2]
					fmt.Println("say ------>" + nickStr + " " + msg_str[1] + toMsgChanStr )
					clt.chatChan <- toMsgChanStr   // 将上线信息传入每个非自己玩家的聊天通道中
				}
			}
		// 玩家的下线通知
		case "offline":
			fmt.Printf("玩家[%s]上线！", msg_str[1])
			for nickStr, clt := range onlineClients {
				if nickStr != msg_str[1] {
					toMsgChanStr := "玩家" + "[" + msg_str[1] + "]" + "已退出聊天室"
					clt.chatChan <- toMsgChanStr   // 将上线信息传入每个非自己玩家的聊天通道中
				}
			}
			// 将退出玩家从在线玩家列表中删除
			delete(onlineClients, msg_str[1])
		}
	}
}
