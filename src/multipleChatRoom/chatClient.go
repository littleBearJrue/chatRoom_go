package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	//打开连接:
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		fmt.Println("Error dialing", err)
		return  // 终止程序
	}
	fmt.Println("connect server successed ... ")

	// 调用结束后关闭socket连接
	defer conn.Close()

	inputReader := bufio.NewReader(os.Stdin)
	fmt.Println("给自己在聊天室起个昵称吧：")
	clientName, _ := inputReader.ReadString('\n')
	trimmedNick := strings.Trim(clientName, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
	fmt.Println("你好，" + trimmedNick + "! 欢迎你加入聊天室")
	// 将玩家登陆信息发送给服务器端
	sendMsgToServer(conn, "online|" + trimmedNick)

	// 创建一个协程不停读取server数据
	go doClientHandle(conn)

	// 给服务器发送信息直到程序退出：
	fmt.Println("你可以开始聊天了，按Q退出聊天室")


	for {
		fmt.Print("please type: ")
		input, _ := inputReader.ReadString('\n')
		trimmedInput := strings.Trim(input, "\r\n")
		sendMsgToServer(conn, "say|" + trimmedNick + "|" + trimmedInput)  //三段字节流 say | 昵称 | 发送的消息
		if trimmedInput == "Q" {
			sendMsgToServer(conn, "offline|" + trimmedNick)  //将quit字节流发送给服务器端
			return
		}
	}
}

// 发送消息给server端
func sendMsgToServer(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("conn write is error, error is: ", err)
	}
}

func doClientHandle(conn net.Conn) {

	for {

		data := make([]byte, 512)  //创建一个字节流
		msg_read, err := conn.Read(data)  //将读取的字节流赋值给msg_read和err
		if msg_read == 0 || err != nil {  //如果字节流为0或者有错误
			break
		}

		fmt.Println("\n" + string(data[0:msg_read]))  //把字节流转换成字符串
	}
}