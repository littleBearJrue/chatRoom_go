package server

import (
	"encoding/json"
	"fmt"
	"os"
)

// 判断文件夹是否存在
func isPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 获取数据储存目录，一旦不存在则创建一个目录
func getDirPath() string {
	dirPath := "./dataBase/"
	iExist, err := isPathExists(dirPath)
	if err != nil {
		fmt.Println("get dir error, error is:", err)
	}
	if iExist {
		return dirPath
	} else {
		err := os.Mkdir(dirPath, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
		}
	}
	return dirPath
}

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
	file,err := os.OpenFile(getDirPath() + fileName, os.O_RDWR|os.O_CREATE, 0101)
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
	fmt.Println("read_chatRoomData-------->", chatRoomData)
	return chatRoomData
}

func InsertChatRoomsDataToFile(fileName string, roomId int, roomName string, users []string) {
	file,fileErr := os.OpenFile(getDirPath() + fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0101)
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

func ReadUserDataFromFile(filename string) map[string]*user{
	buf := make([]byte, 10 * 1024)
	userData := make(map[string]*user)
	file,err := os.OpenFile(getDirPath() + filename, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		fmt.Println("open file error")
	}
	defer file.Close()
	n,_:= file.Read(buf)
	json.Unmarshal(buf[:n],&userData)
	//if jsonErr != nil {
	//	fmt.Println("Json unmarshal is error, error is: ", jsonErr)
	//}
	fmt.Println("read_userData-------->", userData)
	return userData
}


func InsertDataToFile(fileName string, userName string, userPassword string, address string, roomId int, inOnline bool, offTimeStamp int64){
	file,fileErr := os.OpenFile(getDirPath() + fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0766)
	if fileErr != nil {
		fmt.Println("file is not exit!")
	}
	defer file.Close()
	userData[userName] = &user{NickName:userName,Password:userPassword,Address:address,RoomId:roomId, IsOnline:inOnline, OffLineTime:offTimeStamp}
	data, jsonErr := json.Marshal(userData)
	if jsonErr != nil {
		fmt.Println("Json marshal is error, error is: ", jsonErr)
	}
	_, err := file.WriteString(string(data))
	if err != nil {
		fmt.Println("Write file is error, error is: ", err)
	}
}

func ReadChatRecordDataFromFile(filename string) map[string]map[int]map[string] []chatLog{
	buf := make([]byte, 10 * 1024)
	chatRecord := make(map[string]map[int]map[string] []chatLog)
	file,err := os.OpenFile(getDirPath() + filename, os.O_RDWR|os.O_CREATE, 1111)
	if err != nil {
		fmt.Println("open file error")
	}
	defer file.Close()
	n,_:= file.Read(buf)
	json.Unmarshal(buf[:n],&chatRecord)
	//if jsonErr != nil {
	//	fmt.Println("Json unmarshal is error, error is: ", jsonErr)
	//}
	fmt.Println("read_chatRecord-------->", chatRecord)
	return chatRecord
}

func InsertChatRecordToFile(fileName string, chatRecord map[string]map[int]map[string] []chatLog){
	fmt.Println("insert_chatRecord:", chatRecord)
	file,fileErr := os.OpenFile(getDirPath() + fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 1111)
	if fileErr != nil {
		fmt.Println("file is not exit!")
	}
	defer file.Close()
	data, jsonErr := json.Marshal(chatRecord)
	if jsonErr != nil {
		fmt.Println("Json marshal is error, error is: ", jsonErr)
	}
	_, err := file.WriteString(string(data))
	if err != nil {
		fmt.Println("Write file is error, error is: ", err)
	}
}