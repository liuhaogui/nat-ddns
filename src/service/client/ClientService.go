package client

import (
	"bufio"
	"constants"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"log"
	"myproto"
	"net"
	"strconv"
	"strings"
	"time"
	"util/uuid"
	"util/proxy"
	"util/math"
	"entity"
)

type ClientService struct {
}

var id = 0
var flag = true

var sa string
var ck string

var ccs = GetCcsInstance()

var count = 0

func (cs *ClientService) ClientInit(serverAddr string, clientKey string) {
	sa = serverAddr
	ck = clientKey
}

func (cs *ClientService) ClientStart() {
	conn, err := net.Dial("tcp", sa)
	if nil != err {
		log.Println("dial to server:", sa, " err:", err)
		return
	}

	go cs.clientHandle(conn)

	for {
		if cs.IsConnected() {
			time.Sleep(1 * time.Second)
			continue
		} else {
			log.Println("conn is closed!!")
		}
	}
}

func (cs *ClientService) clientHandle(conn net.Conn) {
	cs.cilentAuth(conn)
}

func (cs *ClientService) IsConnected() bool {
	return flag
}

func (cs *ClientService) cilentAuth(conn net.Conn) {
	authMsg := myproto.Msg{
		Id:      proto.Int(id),
		MsgType: proto.Int32(constants.MSG_TYPE_AUTH),
		Key:     proto.String(ck),
		Uri:     proto.String(uuid.GetRandomUUID()),
		Data:    []byte("client_auth"),
	}

	_, err := proxy.MsgWrite(authMsg, conn)
	if nil != err {
		log.Println("client send auth err")
		flag = false
		return
	}

	r := bufio.NewReader(conn)
	for {
		lenBuf := make([]byte, 8)
		r.Read(lenBuf)
		dataLen := math.BytesToInt(lenBuf)
		if dataLen > 0 {
			dataBuf := make([]byte, dataLen)
			n, err := r.Read(dataBuf)
			if nil != err {
				log.Println("read buffer from server err:", err)
				panic(err)
				break
			}
			if n == dataLen {
				msg := myproto.Msg{}
				proto.Unmarshal(dataBuf, &msg)
				msgBytes, _ := json.Marshal(msg)
				log.Println("recv data from server:", string(msgBytes))
				if int(*msg.MsgType) == constants.MSG_TYPE_AUTH {
					id = int(*msg.Id)

					// ping
					go cs.ping(conn)

					dataChan := make(chan myproto.Msg)
					errChan := make(chan error)
					go proxy.ReadWrapper(dataChan, errChan, conn)

					go cs.clientDataProcess(dataChan, errChan, conn)

					break
				}
			}
		}
	}
}

func (cs *ClientService) ping(conn net.Conn) {
	t := time.NewTicker(30 * time.Second)

	heatBeatCount := 0
	heatBeatErrCount := 0

	for {
		select {
		case i := <-t.C:
			heatBeatCount++
			log.Println("ping count:", strconv.Itoa(heatBeatCount), " client ping:", i.Format("2006-01-02 15:04:05"))
			authMsg := myproto.Msg{
				Id:      proto.Int(id),
				MsgType: proto.Int32(constants.MSG_TYPE_HEATBEAT),
				Key:     proto.String(ck),
				Uri:     proto.String(uuid.GetRandomUUID()),
				Data:    []byte(strconv.Itoa(heatBeatCount)),
			}

			_, err := proxy.MsgWrite(authMsg, conn)
			if nil != err {
				heatBeatErrCount++
				log.Println("count:", heatBeatErrCount, "client send heatbeat msg failed!")
			}
		}
	}
}

func (cs *ClientService) clientDataProcess(dataChan chan myproto.Msg, errChan chan error, conn net.Conn) {
	for {
		select {
		case msg := <-dataChan:
			msgBytes, _ := json.Marshal(msg)
			log.Println("recv data from server:", string(msgBytes))

			switch int(*msg.MsgType) {
			case constants.MSG_TYPE_HEATBEAT:
				//heat beat exception process

			case constants.MSG_TYPE_TRANS:
				//handle the trans data
				target, _ := net.Dial("tcp", "www.baidu.com:80")
				str := strings.Replace(string(msg.Data), "127.0.0.1:9191", "www.baidu.com", -1)
				w := bufio.NewWriter(target)
				log.Println("bytes:", str)
				w.Write([]byte(str))
				w.Flush()

				buf := make([]byte, 32*1024)
				written := int64(0)
				for {
					i, err := target.Read(buf)
					if i > 0 {
						log.Println(string(buf[:i]))
						msg := myproto.Msg{
							Id:      msg.Id,
							MsgType: proto.Int(constants.MSG_TYPE_TRANS),
							Key:     msg.Key,
							Uri:     msg.Uri,
							Data:    buf[:i],
						}

						wc, err2 := proxy.MsgWrite(msg, conn)
						written += int64(wc)
						if nil != err2 {
							log.Println("Write Error", err2)
						}
						log.Print(string(buf[:i]))
					}
					if err != nil {
						if err != constants.EOF {
							log.Println("read err:", err)
						}
						break
					}
				}
			}
		case err := <-errChan:
			if nil != err {
				log.Println("An error occured:", err.Error())
				flag = false
				return
			}
		}
	}
}
