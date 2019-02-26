// 总结：
// 1. 凡是需要导出的数据结构，都需要将各个字段写成大写，包括JSON的转换
// 2. int 转 string: 务必使用strconv.Itoa()来进行转换，使用string()可能导致乱码
// 3. int -> string:  strconv.Itoa()    string -> int：  strconv.Atoi()
//
//
//
//
//


package server

import (
	"encoding/json"
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

	PRIVATE_CHAT = "201"
)

const (
	USER_FILE_NAME = "userData.txt"
	CHAT_ROOM_FILE_NAME = "roomData.txt"
	CAHT_CONTENT_FILE_NAME = "chatContentData.txt"
)

// 注意：所有需要导出的结构都需要大写
// 定义聊天室基本数据
type chatRoom struct {
	RoomId int
	RoomName string
	//Clients map[int] client
}

type user struct {
	NickName string
	Password string
	Address string
	IsOnline bool
}

// 定义client存储用户的基本数据
type client struct {
	chatChan chan string
	user
}

// 定义一个map表存储在线的所有玩家信息 key：nickName, client:client实体
var onlineClients = make(map[string] client)

// 定义一个map表储存所有聊天室信息
var chatRooms = make(map[int] chatRoom)

var userData = make(map[string] user)

var heartMsgChan chan string

func Main() {
	fmt.Println("Starting the chat server ...")
	// 优先读取数据库数据
	userData = readUserDataFromFile(USER_FILE_NAME)
	// 更新聊天室数据
	chatRooms = addChatRooms()
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

		go doServerHandle(conn)

		// go heartBreak(conn, 5)

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

		msgLen, err := conn.Read(buf)
		if msgLen == 0 || err != nil {
			fmt.Println("Error reading", err)
			return //终止程序
		}
		// 解析客户端发送过来的数据
		msg_str := strings.Split(string(buf[0:msgLen]), "|")  //将从客户端收到的字节流分段保存到msg_str这个数组中
		// msg_str[0] 存放的数数据类型，包括：“online”,“offline”，“chat”

		fmt.Println("from client msg: ", msg_str)

		switch msg_str[0] {
		// 玩家注册
		case REGISTER:
			var toClientMsg string
			if _,ok := userData[msg_str[1]] ; ok {
				toClientMsg = "用户已存在，请重新注册!"
				fmt.Println(toClientMsg)
			} else {
				// 将新用户数据写入文件中
				insertDataToFile(USER_FILE_NAME, msg_str[1], msg_str[2], clientAddr)
				toClientMsg = "registerSuccess"
			}
			// 传回给客户端
			conn.Write([]byte(toClientMsg + "\n"))
		// 玩家登陆
		case LOGIN:
			var toClientMsg string
			var isSuccess bool = false
			user,ok := userData[msg_str[1]]
			fmt.Println("ok", ok)

			if ok {
				fmt.Println("msg_str[2]", msg_str[2])
				fmt.Println("user.Password", user.Password)
				if user.Password == msg_str[2] {
					toClientMsg = "loginSuccess"
					isSuccess = true
				}else{
					toClientMsg = "密码输入错误!"
				}
			} else {
				toClientMsg = "用户名输入错误!"
			}
			// 传回给客户端
			conn.Write([]byte(toClientMsg + "\n"))

			// 玩家登陆成功
			if isSuccess == true {

			}
		case ROOM_CHOICE:   //选择聊天室
			// 将聊天室列表传给客户端提供选择
			if len(msg_str) == 1 {
				var toClientRoomStr = strconv.Itoa(len(chatRooms)) + "|"
				for i, chatRoom := range chatRooms{
					// 这里int 转 string 务必使用strconv.Itoa()，使用string()会出现乱码
					roomName := strconv.Itoa(i) + "." + chatRoom.RoomName + "\n"
					toClientRoomStr = toClientRoomStr + roomName
				}
				fmt.Println("ROOM_CHOICE", toClientRoomStr)
				// 传回给客户端
				conn.Write([]byte(toClientRoomStr + "\n"))
			} else {
				// 根据玩家选择的聊天室进入对于聊天室展开对话
				index,_:= strconv.Atoi(msg_str[1])
				curRoomName := chatRooms[index].RoomName
				// 获取该聊天室的总人数
				toClientMsg := "欢迎你进入" + curRoomName + "聊天室！此聊天室总人数为100人"

				// 传回给客户端
				conn.Write([]byte(toClientMsg + "\n"))
			}

		case ONLINE:  // 玩家登陆上线
			clt := client{make(chan string), user{msg_str[1], clientAddr, clientAddr, true}}
			onlineClients[msg_str[1]] = clt

			go sendMsgToClient(clt, conn)

			fmt.Printf("玩家[%s]上线！\n", msg_str[1])
			for nickStr, clt := range onlineClients {
				if nickStr != msg_str[1] {
					toMsgChanStr := "玩家" + "[" + msg_str[1] + "]" + "已上线，你们可以欢快的聊天了"
					clt.chatChan <- toMsgChanStr   // 将上线信息传入每个非自己玩家的聊天通道中
				}
			}
		case CHAT:  // 玩家的聊天内容，转发给客户端
			for nickStr, clt := range onlineClients {
				if nickStr != msg_str[1] {
					toMsgChanStr := "[" + msg_str[1] + "]： " + msg_str[2]
					clt.chatChan <- toMsgChanStr   // 将上线信息传入每个非自己玩家的聊天通道中
				}
			}
		case P_CHAT:   //私聊具体内容 msg_str[1]:私聊的玩家   msg_str[2]：发送信息的玩家 msg_str[3]：聊天的具体内容
			//var toClientMsg string = P_CHAT + "|"
			//if len(msg_str) == 1 {   // client端只输入“@”调起所有用户列表
			//	var allUserName string
			//	for name, _ := range userData {
			//		allUserName = allUserName + " " + name
			//	}
			//	toClientMsg = toClientMsg + allUserName
			//
			//} else {   // client端确定私聊用户 “@” + 用户名
			//	p_name := msg_str[1]
			//	if _,ok := userData[p_name] ; ok {
			//		toClientMsg = toClientMsg + "success"
			//	}
			//}
			//fmt.Println("p_char", toClientMsg)
			//// 传回给客户端
			//conn.Write([]byte(toClientMsg + "\n"))

			// 假如输入的私聊玩家是自己，则提醒客户端
			var toClientMsg string
			if msg_str[1] == msg_str[2] {
				toClientMsg = "不能跟自己私聊！"
			} else {
				toClientMsg = "[" + msg_str[2] + "]： " + msg_str[3]
			}
			for nickStr, clt := range onlineClients {
				if nickStr == msg_str[1] {
					clt.chatChan <- toClientMsg   // 将上线信息传入每个非自己玩家的聊天通道中
				}
			}
		case PRIVATE_CHAT:

		case OFFLINE:  // 玩家的下线通知
			fmt.Printf("玩家[%s]下线！'\n'", msg_str[1])
			for nickStr, clt := range onlineClients {
				if nickStr != msg_str[1] {
					toMsgChanStr := "玩家" + "[" + msg_str[1] + "]" + "已退出聊天室"
					clt.chatChan <- toMsgChanStr   // 将上线信息传入每个非自己玩家的聊天通道中
				}
			}
			// 将退出玩家从在线玩家列表中删除
			delete(onlineClients, msg_str[1])

		case HEART:  // 心跳包
			fmt.Println("heartBeat Msg ----->", msg_str[1])
			heartMsgChan <- msg_str[1]
		}
	}
}

// 转发用户的数据给其他用户
func sendMsgToClient(clt client, conn net.Conn){
	// 循环遍历每个玩家，将玩家的储存的通道数据一个个通过conn.write转给客户端

	for {
		for msgInfo := range clt.chatChan {
			fmt.Println("write -----> ", clt.NickName, msgInfo )
			_, err := conn.Write([]byte(msgInfo + "\n"))
			if err != nil {
				fmt.Println("conn write is error, error is: ", err)
			}
		}
	}
}

// 协程检测心跳包
func heartBreak(conn net.Conn, timeout int) {
	//fmt.Println("heartBreak---->", <- heartMsgChan)
	select {
	case <- heartMsgChan:
		fmt.Println("heart break from client: ", <- heartMsgChan)
		err := conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		if err != nil {
			fmt.Println("conn setDeadLine is error")
		}
	case <- time.After(time.Second * 5):
		fmt.Println(conn.RemoteAddr().String(), "time out")
		conn.Close()
	}
}


func readUserDataFromFile(filename string) map[string]user{
	buf := make([]byte, 10 * 1024)
	userData := make(map[string]user)
	file,err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		fmt.Println("open file error")
	}
	defer file.Close()
	n,_:= file.Read(buf)
	json.Unmarshal(buf[:n],&userData)
	return userData
}


func insertDataToFile(fileName string, userName string, userPassword string, address string){
	file,err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		fmt.Println("file is not exit!")
	}
	defer file.Close()
	userData[userName] = user{userName,userPassword,address, true }
	fmt.Println("insertDataToFile", userData)
	data, _ := json.Marshal(userData)
	fmt.Println("json", data)
	fmt.Println("stringFromJson", string(data))
	file.WriteString(string(data))
}

func addChatRooms() map[int] chatRoom{
	chatRoomName := []string{"天蝎座", "天秤座", "金羊座", "摩羯座", "处女座"}
	rooms := make(map[int] chatRoom)
	for i := 0; i < len(chatRoomName); i++ {
		rooms[i] = chatRoom{i, chatRoomName[i]}
	}
	return rooms
}
