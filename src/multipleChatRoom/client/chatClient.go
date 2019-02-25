package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)
// 定义协议码
// 1. 101 :玩家注册登陆
// 2. 102：玩家上线
// 3. 103：玩家聊天内容
// 4. 104：玩家下线
//
const (
	// 核心
	HEART = "100"
	LOGIN = "101"
	REGISTER = "102"
	ONLINE = "103"
	CHAT = "104"
	OFFLINE = "105"

	// 对聊天方式标示
	P_CHAT = "@"     // 私聊标示
	HINT_CHAT = "#"  // 常用语
	Exit = "Q"

	PRIVATE_CHAT = "201"
)

// 接收数据的通道
var recvChan = make(chan string)
// 发送数据的通道
var sendChan = make(chan string)
// 用户名
var userName string
// 用户密码
var userPassword string

func Main() {
	//打开连接:
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		fmt.Println("Error dialing", err)
		return  // 终止程序
	}
	fmt.Println("connect server successed ... ")

	// 启用心跳包
	// go heartBreakHandle(conn)

	// 调用结束后关闭socket连接
	defer conn.Close()

	// 创建一个协程不停读取conn的数据写进recvChan通道中
	go doClientRecvData(conn)
	// 创建一个协程不停的从recvChan通道中读取数据写进conn中
	go doClientSendData(conn)


	fmt.Println("连接成功！请进行以下操作：1、登录  2、注册")
	inputReader := bufio.NewReader(os.Stdin)
	userChoice, _ := inputReader.ReadString('\n')
	choiceInput := strings.Trim(userChoice, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
	// 玩家登陆操作
	if choiceInput == "1" {
		userLogin(inputReader)
	// 玩家注册操作
	}else if choiceInput == "2" {
		userRegister(inputReader)
	}else {
		fmt.Println("输入错误！")
		return
	}
	// 开启一个线程显示获取到server数据
	go displayMsgContent()

	// 主线程上不断循环获取用户的聊天内容
	for {
		fmt.Print("please type: ")
		input, _ := inputReader.ReadString('\n')
		trimmedInput := strings.Trim(input, "\r\n")
		switch trimmedInput {
		// 进入私聊
		case P_CHAT:
			//// 发送给server
			//sendChan <- P_CHAT
			//
			//// 收到server的回答
			//fmt.Println("len of chan:", len(recvChan))
			//result := <- recvChan
			//trimResult := strings.Trim(result, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
			//fmt.Println("trimResult", trimResult)
			//fmt.Println("请输入你要私聊的用户: ")
			//name, _ := inputReader.ReadString('\n')
			//trimmedName := strings.Trim(name, "\r\n")
			//
			//sendChan <- P_CHAT + "|" + trimmedName
			//
			//// 从接收通道中读取服务器数据，得到登陆结果
			//response := <- recvChan
			//trimmedResponse := strings.Trim(response, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
			//// 输入正确且在线存在这个人
			//if trimmedResponse == "success" {
			//	fmt.Print("please type with " + trimmedName + ": ")
			//	p_input, _ := inputReader.ReadString('\n')
			//	p_trimmedInput := strings.Trim(p_input, "\r\n")
			//	sendChan <- PRIVATE_CHAT + "|" + userName + "|" + p_trimmedInput  //三段字节流 say | 昵称 | 发送的消息
			//
			//} else {
			//	// 输入有误，重新回到聊天界面
			//	fmt.Println("请检查是否输入的私聊玩家有误！")
			//	break
			//}
			fmt.Println("请输入你要私聊的用户: ")
			name, _ := inputReader.ReadString('\n')
			trimmedName := strings.Trim(name, "\r\n")
			fmt.Println("请输入你要私聊的内容: ")
			content, _ := inputReader.ReadString('\n')
			trimmedContent:= strings.Trim(content, "\r\n")
			sendChan <- P_CHAT + "|" + trimmedName + "|" + userName + "|" + trimmedContent
		case HINT_CHAT:
			// 常用语提示


		case Exit:
			// 退出聊天
			sendChan <- OFFLINE + "|" + userName  //将quit字节流发送给服务器端
			return
		default:
			// 默认群发
			sendChan <- CHAT + "|" + userName + "|" + trimmedInput  //三段字节流 say | 昵称 | 发送的消息
		}
	}
}

func userLogin(inputReader *bufio.Reader ) {
	LOOP: for {
		fmt.Println("请输入用户名：")
		name, _ := inputReader.ReadString('\n')
		trimName := strings.Trim(name, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
		fmt.Println("请输入密码：")
		password, _ := inputReader.ReadString('\n')
		trimPassword := strings.Trim(password, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"

		// 将用户登陆的数据放入发送通道中上传到服务器
		sendChan <- LOGIN + "|" + trimName + "|" + trimPassword

		// 从接收通道中读取服务器数据，得到登陆结果
		result := <- recvChan
		trimResult := strings.Trim(result, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
		if trimResult == "loginSuccess" {

			userName = trimName
			userPassword = trimPassword

			fmt.Println("你好，" + userName + "! 欢迎你加入聊天室")
			// 将玩家登陆信息发送给服务器端
			sendChan <- ONLINE + "|" + userName

			fmt.Println("你可以开始聊天了，按Q退出聊天室")

			break
		} else {
			// 打印错误消息
			fmt.Println(trimResult)
			// 登陆失败则重新启动登陆
			goto LOOP
		}
	}

}

func userRegister(inputReader *bufio.Reader) {
	LOOP: for {
		fmt.Println("请输入用户名：")
		name, _ := inputReader.ReadString('\n')
		trimName := strings.Trim(name, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
		fmt.Println("请输入密码：")
		password, _ := inputReader.ReadString('\n')
		trimPassword := strings.Trim(password, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"

		// 将用户注册的数据放入发送通道中上传到服务器
		sendChan <- REGISTER + "|" + trimName + "|" + trimPassword

		// 从接收通道中读取服务器数据，得到注册结果
		result := <- recvChan
		trimResult := strings.Trim(result, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
		if trimResult == "registerSuccess" {
			fmt.Println("恭喜你注册成功，请完成登录")
			// 注册成功，转去登陆
			userLogin(inputReader)
			break
		}else{
			fmt.Println(trimResult)
			goto LOOP
		}
	}
}


// 每2s发送一个心跳包
//func heartBreakHandle(conn net.Conn) {
//	for {
//		fmt.Println("heartBreakHandle.....")
//		 heart_word := "heart break ... "
//		 sendMsgToServer(conn, HEART + "|" + heart_word)
//		time.Sleep(2 * time.Second)
//	}
//}

// 从发送数据通道中将数据取出来，通过conn.write写进去传给server
func doClientSendData(conn net.Conn) {
 for {
 	msg := <- sendChan
	 _, err := conn.Write([]byte(msg))
	 if err != nil {
		 fmt.Println("conn write is error, error is: ", err)
	 }
 }
}

// 将从conn.read读出的数据写进接收通道中，等待输出显示
func doClientRecvData(conn net.Conn) {
	for {
		buf := make([]byte, 512)  //创建一个字节流
		msgLen, err := conn.Read(buf)  //将读取的字节流赋值给msg_read和err
		if msgLen == 0 || err != nil {  //如果字节流为0或者有错误
			break
		}

		recvChan <- string(buf[:msgLen])
	}
}

func displayMsgContent() {
	for {
		msg :=<- recvChan
		fmt.Println("\n" + "from ", string(msg))  //把字节流转换成字符串
		fmt.Print("please type: ")
	}
}

