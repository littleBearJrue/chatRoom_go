package server

import (
	"encoding/json"
	"fmt"
	"os"
)

func initChatRooms() map[int]*chatRoom{
	chatRoomName := []string{"天蝎座", "天秤座", "金羊座", "摩羯座", "处女座"}
	rooms := make(map[int] *chatRoom)
	for i := 0; i < len(chatRoomName); i++ {
		rooms[i] = &chatRoom{RoomId:i, RoomName:chatRoomName[i], Users:[]string{}, clients: map[string] client{}}
	}
	return rooms
}

func ReadChatDataFromFile(fileName string) map[int]*chatRoom {
	buf := make([]byte, 10 * 1024)
	chatRoomData := make(map[int]*chatRoom)
	file,err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0101)
	if err != nil {
		fmt.Println("open file error")
	}
	defer file.Close()
	n,_:= file.Read(buf)
	json.Unmarshal(buf[:n],&chatRoomData)
	//if jsonErr != nil {
	//	fmt.Println("Json unmarshal is error, error is: ", jsonErr)
	//}
	// 如果没在保存的文件中读取到数据，则优先使用默认的数据
	if len(chatRoomData) == 0 {
		chatRoomData = initChatRooms()
	}
	return chatRoomData
}

func InsertChatRoomsDataToFile(fileName string, roomId int, roomName string, users []string) {
	file,fileErr := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0101)
	if fileErr != nil {
		fmt.Println("file is not exit!")
	}
	defer file.Close()
	chatRooms[roomId] = &chatRoom{RoomId:roomId, RoomName:roomName, Users:users}
	data, jsonErr := json.Marshal(chatRooms)
	if jsonErr != nil {
		fmt.Println("Json marshal is error, error is: ", jsonErr)
	}
	_, err := file.WriteString(string(data))
	if err != nil {
		fmt.Println("Write file is error, error is: ", err)
	}
}

func ReadUserDataFromFile(filename string) map[string]user{
	buf := make([]byte, 10 * 1024)
	userData := make(map[string]user)
	file,err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		fmt.Println("open file error")
	}
	defer file.Close()
	n,_:= file.Read(buf)
	json.Unmarshal(buf[:n],&userData)
	//if jsonErr != nil {
	//	fmt.Println("Json unmarshal is error, error is: ", jsonErr)
	//}
	fmt.Println("userData-------->", userData)
	return userData
}


func InsertDataToFile(fileName string, userName string, userPassword string, address string, roomId int){
	file,fileErr := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0766)
	if fileErr != nil {
		fmt.Println("file is not exit!")
	}
	defer file.Close()
	userData[userName] = user{NickName:userName,Password:userPassword,Address:address, RoomId:roomId}
	data, jsonErr := json.Marshal(userData)
	if jsonErr != nil {
		fmt.Println("Json marshal is error, error is: ", jsonErr)
	}
	_, err := file.WriteString(string(data))
	if err != nil {
		fmt.Println("Write file is error, error is: ", err)
	}
}