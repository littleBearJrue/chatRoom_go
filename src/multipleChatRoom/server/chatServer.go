// 总结：
// 1. 凡是需要导出的数据结构，都需要将各个字段写成大写，包括JSON的转换
// 2. int 转 string: 务必使用strconv.Itoa()来进行转换，使用string()可能导致乱码
// 3. int -> string:  strconv.Itoa()    string -> int：  strconv.Atoi()
// 4. 遇到map表下存储数组（切片）的情况，直接将已填充好数据的数组塞进map对应的字段中，会存在报错提示：chatRooms[index].users = []string{"aaaa", "bbb"}，这时候就会报错。在初始化时应该为使用地址传递的方式。var chatRooms = make(map[int]*chatRoom)
// 5. 在多次读写文件时，发现转为json格式存入文件总是莫名其妙多了个“}”,导致读取文件的时候，因json格式错误，而读取不出来。这时候通过增加os.O_TRUNC类型的方式解决，既每次写文件都清空当前文件内容
//
//
//
//


package server

import (
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
	CHAT_ROOM_FILE_NAME = "roomData.txt"
	USER_FILE_NAME = "userData.txt"
)

// 注意：所有需要导出的结构都需要大写
// 定义聊天室基本数据
type chatRoom struct {
	RoomId int          // 房间id
	RoomName string     // 房间名
	Users []string      // 已登录过此房间的用户，这里保存用户名，映射用户表，可获取用户的具体数据
	clients map[string] client // 持有每个客户端的连接
}

// 定义client存储用户的基本数据
type user struct {
	NickName string    // 用户昵称
	Password string    // 用户密码
	Address string    // 用户ip地址
	RoomId int        // 用户所在房间id
	IsOnline bool     // 用户是否在线
	ContentRecord map[int]map[string] []chatLog   // 聊天记录 key1: roomId, key2:"all" or "someone", value: chatLog ==> 某个房间内收到的某个人或者所有人的聊天记录
}

type chatLog struct {
	chatTime int32   //聊天时间节点
	content []string  // 具体聊天内容
}

type client struct {
	chatChan chan string
	userName string
}

// 定义一个map表储存所有聊天室信息
var chatRooms = make(map[int]*chatRoom)

// 定义用户数据
var userData = make(map[string] user)

// 聊天记录数据
var chatHistory = make( map[int]map[string] []chatLog)

var heartMsgChan chan string

func Main() {
	fmt.Println("Starting the chat server ...")

	// 优先读取房间数据库数据,如果没有，则初始化一份房间数据表
	chatRooms = ReadChatDataFromFile(CHAT_ROOM_FILE_NAME)
	// 再读取用户相关数据库
	userData = ReadUserDataFromFile(USER_FILE_NAME)
	// 最后再读取玩家的聊天记录，查看是否存在离线记录
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
				// TODO 注册成功将新用户数据写入文件中，此时因为用户尚未进入聊天室，尚未保存roomId
				InsertDataToFile(USER_FILE_NAME, msg_str[1], msg_str[2], clientAddr, -1)

				fmt.Println("userData----->", userData)

				toClientMsg = "registerSuccess"
			}
			// 传回给客户端
			sendMsgToSelf(toClientMsg + "\n", conn)
		// 玩家登陆
		case LOGIN:
			var toClientMsg string
			user,ok := userData[msg_str[1]]
			if ok {
				if user.Password == msg_str[2] {
					toClientMsg = "loginSuccess"
				}else{
					toClientMsg = "密码输入错误!"
				}
			} else {
				toClientMsg = "用户名输入错误!"
			}
			// 传回给客户端
			sendMsgToSelf(toClientMsg + "\n", conn)
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
				sendMsgToSelf(toClientRoomStr + "\n", conn)
			} else {

				// 根据玩家选择的聊天室进入对于聊天室展开对话
				index,_:= strconv.Atoi(msg_str[2])
				curRoomName := chatRooms[index].RoomName

				// 进入聊天室成功，保存玩家数据,写入房间id
				InsertDataToFile(USER_FILE_NAME, userData[msg_str[1]].NickName, userData[msg_str[1]].Password, userData[msg_str[1]].Address, index)

				// 写入成功登录之后的连接对象map
				var onlineClients = make(map[string] client)
				if chatRooms[index].clients != nil && len(chatRooms[index].clients) > 0{
					onlineClients = chatRooms[index].clients
				}
				clt := client{make(chan string), msg_str[1]}
				onlineClients[msg_str[1]] = clt


				// 将房间数据保存到文件中
				var isInsert bool = true
				for _, name := range chatRooms[index].Users {
					if name == msg_str[1] {
						isInsert = false
					}
				}
				// 不存在的时候才插入
				if isInsert {
					chatRooms[index].Users = append(chatRooms[index].Users, msg_str[1])
				}

				// 不保存clients字段
				InsertChatRoomsDataToFile(CHAT_ROOM_FILE_NAME, index, chatRooms[index].RoomName, chatRooms[index].Users)

				chatRooms[index].clients = onlineClients

				// 获取该聊天室的总人数
				toClientMsg := "欢迎你进入" + curRoomName + "聊天室！此聊天室总人数为" + strconv.Itoa(len(chatRooms[index].Users)) + "人"

				// 传回给客户端
				sendMsgToSelf(toClientMsg + "\n", conn)

				go sendMsgToOthers(clt, conn)
			}
		case ONLINE:  // 玩家登陆上线

			sendMsgToSelf("这是模拟上线接收到的离线消息： [111]: 你好" + "\n", conn)

			fmt.Printf("玩家[%s]上线！\n", msg_str[1])
			curRoomId := userData[msg_str[1]].RoomId
			for nickStr, clt := range chatRooms[curRoomId].clients {
				if nickStr != msg_str[1] {
					toMsgChanStr := "玩家" + "[" + msg_str[1] + "]" + "已上线，你们可以欢快的聊天了"
					clt.chatChan <- toMsgChanStr   // 将上线信息传入每个非自己玩家的聊天通道中
				}
			}
		case CHAT:  // 玩家的聊天内容，转发给客户端
			// ContentRecord map[int]map[string] []chatLog   // 聊天记录 key1: roomId, key2:"all" or "someone", value: chatLog ==> 某个房间内收到某个人或者所有人的聊天记录



			curRoomId := userData[msg_str[1]].RoomId
			for nickStr, clt := range chatRooms[curRoomId].clients {
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
			curRoomId := userData[msg_str[2]].RoomId
			for nickStr, clt := range chatRooms[curRoomId].clients {
				if nickStr == msg_str[1] {
					clt.chatChan <- toClientMsg   // 将上线信息传入每个非自己玩家的聊天通道中
				}
			}
		case PRIVATE_CHAT:

		case OFFLINE:  // 玩家的下线通知
			fmt.Printf("玩家[%s]下线！'\n'", msg_str[1])
			curRoomId := userData[msg_str[1]].RoomId
			for nickStr, clt := range chatRooms[curRoomId].clients {
				if nickStr != msg_str[1] {
					toMsgChanStr := "玩家" + "[" + msg_str[1] + "]" + "已退出聊天室"
					clt.chatChan <- toMsgChanStr   // 将上线信息传入每个非自己玩家的聊天通道中
				}
			}
			// 将退出玩家从在线玩家列表中删除
			delete(chatRooms[curRoomId].clients, msg_str[1])

		case HEART:  // 心跳包
			fmt.Println("heartBeat Msg ----->", msg_str[1])
			heartMsgChan <- msg_str[1]
		}
	}
}

// 转发信息给自己，主要是那些操作是否成功，以及相关提示语
func sendMsgToSelf(toClientMsg string, conn net.Conn) {
	_, err := conn.Write([]byte(toClientMsg + "\n"))
	if err != nil {
		fmt.Println("conn write to self is error, error is: ", err)
	}
}

// 转发用户的数据给其他用户
func sendMsgToOthers(clt client, conn net.Conn){
	// 循环遍历每个玩家，将玩家的储存的通道数据一个个通过conn.write转给客户端

	for {
		for msgInfo := range clt.chatChan {
			fmt.Println("write -----> ", clt.userName, msgInfo )
			_, err := conn.Write([]byte(msgInfo + "\n"))
			if err != nil {
				fmt.Println("conn write to others is error, error is: ", err)
				// 在这里储存玩家的离线聊天记录
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

