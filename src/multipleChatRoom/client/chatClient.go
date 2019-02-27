package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
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
	ROOM_CHOICE = "103"
	ONLINE = "104"
	CHAT = "105"
	OFFLINE = "106"

	// 对聊天方式标示
	P_CHAT = "@"     // 私聊标示
	HINT_CHAT = "#"  // 常用语
	Exit = "Q"

	PRIVATE_CHAT = "201"
)

// 最大房间数
var MAX_ROMM_NUM = 5

// 心跳包时间间隔
var HEART_BEAT_UNIT int = 3

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

	// 调用结束后关闭socket连接
	defer conn.Close()

	// 创建一个协程不停读取conn的数据写进recvChan通道中
	go doClientRecvData(conn)
	// 创建一个协程不停的从recvChan通道中读取数据写进conn中
	go doClientSendData(conn)

	fmt.Println("连接成功！请进行以下操作：1、登录  2、注册")
	inputReader := bufio.NewReader(os.Stdin)
	LOOP :{
		userChoice, _ := inputReader.ReadString('\n')
		choiceInput := strings.Trim(userChoice, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
		// 玩家登陆操作
		if choiceInput == "1" {
			userLogin(inputReader)
			// 玩家注册操作
		}else if choiceInput == "2" {
			userRegister(inputReader)
		}else {
			fmt.Println("输入错误！请重新输入: ")
			goto LOOP
		}
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
			time.Sleep(1 * time.Second)   // 延时1s执行退出
			return
		default:
			// 默认群发
			sendChan <- CHAT + "|" + userName + "|" + trimmedInput  //三段字节流 say | 昵称 | 发送的消息
		}
	}
}

// 玩家登陆
func userLogin(inputReader *bufio.Reader ) {
	LOOP: for {
		fmt.Println("请输入用户名：")
		name, _ := inputReader.ReadString('\n')
		trimmedName := strings.Trim(name, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
		fmt.Println("请输入密码：")
		password, _ := inputReader.ReadString('\n')
		trimmedPassword := strings.Trim(password, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"

		// 将用户登陆的数据放入发送通道中上传到服务器
		sendChan <- LOGIN + "|" + trimmedName + "|" + trimmedPassword

		// 从接收通道中读取服务器数据，得到登陆结果
		result := <- recvChan
		trimmedResult := strings.Trim(result, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
		if trimmedResult == "loginSuccess" {
			userName = trimmedName
			userPassword = trimmedPassword

			// 收到目前已有的聊天室供玩家选择
			fmt.Println("恭喜你登陆成功，请先选择你要加入的聊天室：")
			// 往通道中放入请求房间列表的数据
			sendChan <- ROOM_CHOICE

			// 从接收通道中读取服务器数据，得到登陆结果
			response := <- recvChan
			trimmedResponse := strings.Trim(response, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
			responseStr := strings.Split(trimmedResponse, "|")
			MAX_ROMM_NUM, _ = strconv.Atoi(responseStr[0])
			// 根据后端传过来的聊天室信息展示
			fmt.Println(responseStr[1])
			var trimmedRoomChoice string
			RECHOICE: {
				fmt.Print("请输入：")
				roomChoice, _ := inputReader.ReadString('\n')
				trimmedRoomChoice = strings.Trim(roomChoice, "\r\n") // Windows 平台下用 "\r\n"，Linux平台下使用 "\n"
				index, _ := strconv.Atoi(trimmedRoomChoice)
				if index >= MAX_ROMM_NUM {
					fmt.Println("选择有误！请重新输入：")
					goto RECHOICE
				}
			}
			// 通知server玩家进入的聊天室
			sendChan <- ONLINE + "|" + userName + "|" + trimmedRoomChoice

			// 从接收通道中读取服务器数据，得到登陆结果
			enterRoomMsg := <- recvChan
			// 展示server端发送过来的数据
			fmt.Println(enterRoomMsg)
			fmt.Println("你可以开始聊天了，按Q退出聊天室")

			// 登陆成功启用心跳包
			go heartBreakHandle(userName)

			break
		} else {
			// 打印错误消息
			fmt.Println(trimmedResult)
			// 登陆失败则重新启动登陆
			goto LOOP
		}
	}

}

// 玩家注册
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

// 每3s发送一个心跳包
func heartBreakHandle(userName string) {
	for {
		// fmt.Println("heartBreakHandle.....")
		 heartMsg := "heart break ... "
		sendChan <- HEART + "|" + userName + "|" + heartMsg
		time.Sleep(3 * time.Second)
	}
}

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
		fmt.Println("\n" +  string(msg))  //把字节流转换成字符串
		fmt.Print("please type: ")
	}
}


