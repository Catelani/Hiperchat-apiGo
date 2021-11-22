package libs

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/streadway/amqp"

	"github.com/Mateus-pilo/go-whats-opt/hlp"
	whatsapp "github.com/Rhymen/go-whatsapp"
	qrcode "github.com/skip2/go-qrcode"
)

var wac = make(map[string]*whatsapp.Conn)
var amq = ConnectionMqp()



type waHandler struct{
	c *whatsapp.Conn
	jid string
	created uint64
}


type msgResponse struct {
	whatsapp.TextMessage
	Jid string `json:"jid"`
}

type msgResponseImage struct {
	whatsapp.ImageMessage
	Type  string
	Jid string `json:"jid"`
	Path string `json:"file"`
	File_name string `json:"file_name"`
}

type msgResponseDocument struct {
	whatsapp.DocumentMessage
	Type  string
	Jid string `json:"jid"`
	Path string `json:"file"`
	File_name string `json:"file_name"`
}


type msgResponseVideo struct {
	whatsapp.VideoMessage
	Type string
	Jid string `json:"jid"`
	Path string `json:"file"`
	File_name string `json:"file_name"`
}

type msgResponseAudio struct {
	whatsapp.AudioMessage
	Type string
	Jid string `json:"jid"`
	Path string `json:"file"`
	File_name string `json:"file_name"`
}

type responseContacts struct {
	Contacts []whatsapp.Contact `json:"contacts"`
	Jid string `json:"jid_company"`
}

type deviceInfo struct {
	Info []interface{}
	Jid string `json:"jid_company"`
	PhoneNumber string `json:"number_whats"`
	Battery int `json:"battery"`
	Wifi bool `json:"wifi"`
	Charge bool `json:"charge"`
	Is24h bool `json:"is24h"`
	Platform string `json:"platform"`

}

type errorDisconnected struct {
	Jid string `json:"jid_company"`
}



func (h *waHandler) HandleError(err error) {

	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: [JID: "+h.jid+" ] %v", e.Err)
		log.Println("Waiting 30sec...")
		notifyError(h.jid)
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")

		
		file := hlp.Config.GetString("SERVER_STORE_PATH") + "/" + h.jid + ".gob"

		qrstr := make(chan string)
		errmsg := make(chan error)

		go func() {
			WASessionConnect(h.jid, 5, file, qrstr, errmsg)
		}()
		select {
			case err := <-errmsg:
				if len(err.Error()) != 0 {
					log.Println("Reconnect Failed: [Jid: "+h.jid+"]")
					return
				}	
		}	
	}
}


func (h *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	if message.Info.FromMe == false && message.Info.Timestamp >= h.created  {
		
		responseMessage := msgResponse{TextMessage: message}
		responseMessage.Jid = h.jid
		jsonStr, _ := json.Marshal(responseMessage)
		
		err := amq.Channel.Publish(
			"",     // exchange
			"msgSend", // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing {
				ContentType: "text/plain",
				Body:        []byte(jsonStr),
			})
		
		if err != nil {
			log.Println("Fail Publish Msg [Jid: "+h.jid+"]")
		}
	}
}

func (h *waHandler) HandleImageMessage(message whatsapp.ImageMessage) { 
	
	
	if message.Info.FromMe == false && message.Info.Timestamp >= h.created {
		data, err := message.Download()
		if err != nil {
			if err != whatsapp.ErrMediaDownloadFailedWith410 {
				return
			}
			data, err = message.Download()
			if err != nil {
				return
			
			}
		}
		filename := fmt.Sprintf("%v/%v.%v", "/var/whats/image", message.Info.Id, strings.Split(message.Type, "/")[1])
		file, err := os.Create(filename)
		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			fmt.Printf("[!] %v\n", err)
			return
		}

		responseMessage := msgResponseImage{ImageMessage: message, Type: "image", Jid: h.jid, Path: filename, File_name: filepath.Base(filename) }
		

		var inInterface map[string]string
		jsonStr, _ := json.Marshal(responseMessage)
		json.Unmarshal(jsonStr, &inInterface)
		

		err = amq.Channel.Publish(
			"",     // exchange
			"msgSend", // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing {
				DeliveryMode: amqp.Persistent,
				ContentType: "text/plain",
				Body: []byte(jsonStr),
			})
		
		if err != nil {
			log.Println("Fail Publish Msg [Jid: "+h.jid+"]")
		}

	}
	
}

func (h *waHandler) HandleDocumentMessage(message whatsapp.DocumentMessage) {
	
	if message.Info.FromMe == false  && message.Info.Timestamp >= h.created {
	
		data, err := message.Download()
		if err != nil {
			if err != whatsapp.ErrMediaDownloadFailedWith410 {
				return
			}
			data, err = message.Download()
			if err != nil {
				return
			
			}
		}
		filename := fmt.Sprintf("%v/%v.%v", "/var/whats/document", message.Info.Id, strings.Split(message.Type, "/")[1])
		file, err := os.Create(filename)
		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			fmt.Printf("[!] %v\n", err)
			return
		}
	
		responseMessage := msgResponseDocument{DocumentMessage: message, Type: "file", Jid: h.jid, Path:	filename, File_name: filepath.Base(filename) }
		

		var inInterface map[string]string
		jsonStr, _ := json.Marshal(responseMessage)
		json.Unmarshal(jsonStr, &inInterface)
		
  
		err = amq.Channel.Publish(
			"",     // exchange
			"msgSend", // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing {
				DeliveryMode: amqp.Persistent,
				ContentType: "text/plain",
				Body: []byte(jsonStr),
			})
		
		if err != nil {
			log.Println("Fail Publish Msg [Jid: "+h.jid+"]")
		}
	}
}

func (h *waHandler) HandleVideoMessage(message whatsapp.VideoMessage) {
	
	if message.Info.FromMe == false  && message.Info.Timestamp >= h.created {
	
		data, err := message.Download()
		if err != nil {
			if err != whatsapp.ErrMediaDownloadFailedWith410 {
				return
			}
			data, err = message.Download()
			if err != nil {
				return
			
			}
		}
		filename := fmt.Sprintf("%v/%v.%v", "/var/whats/video", message.Info.Id, strings.Split(message.Type, "/")[1])
		file, err := os.Create(filename)
		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			fmt.Printf("[!] %v\n", err)
			return
		}

		responseMessage := msgResponseVideo{VideoMessage: message, Type: "video", Jid: h.jid, Path: filename, File_name:filepath.Base(filename) }
		

		var inInterface map[string]string
		jsonStr, _ := json.Marshal(responseMessage)
		json.Unmarshal(jsonStr, &inInterface)
		
		err = amq.Channel.Publish(
			"",     // exchange
			"msgSend", // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing {
				DeliveryMode: amqp.Persistent,
				ContentType: "text/plain",
				Body: []byte(jsonStr),
			})
		
		if err != nil {
			log.Println("Fail Publish Msg [Jid: "+h.jid+"]")
		}
	}
}

func (h *waHandler) HandleAudioMessage(message whatsapp.AudioMessage){	
	
	if message.Info.FromMe == false  && message.Info.Timestamp >= h.created {
	
		data, err := message.Download()
		if err != nil {
			if err != whatsapp.ErrMediaDownloadFailedWith410 {
				return
			}
			data, err = message.Download()
			if err != nil {
				return
			
			}
		}
		filename := fmt.Sprintf("%v/%v.%v", "/var/whats/audio", message.Info.Id, strings.Split(message.Type, "/")[1])
		file, err := os.Create(filename)
		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			fmt.Printf("[!] %v\n", err)
			return
		}

		responseMessage := msgResponseAudio{AudioMessage: message, Type: "audio", Jid: h.jid, Path:filename, File_name:  filepath.Base(filename)}
		

		var inInterface map[string]string
		jsonStr, _ := json.Marshal(responseMessage)
		json.Unmarshal(jsonStr, &inInterface)

		err = amq.Channel.Publish(
			"",     // exchange
			"msgSend", // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing {
				DeliveryMode: amqp.Persistent,
				ContentType: "text/plain",
				Body: []byte(jsonStr),
			})
		
		if err != nil {
			log.Println("Fail Publish Msg [Jid: "+h.jid+"]")
		}

	}
}


func (h *waHandler) HandleContactList(Contacts[] whatsapp.Contact){
	responseContact := responseContacts{Contacts: Contacts}
	responseContact.Jid = h.jid
	
	jsonStr, _ := json.Marshal(responseContact)
	urlPost := hlp.Config.GetString("SERVER_API_NODE_CONTACTS")
	req, _ := http.NewRequest("POST", urlPost, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		hlp.LogPrintln("error", "http-server", err.Error())
		return
	}


	defer resp.Body.Close()	
}

func (h *waHandler) HandleChatList(contacts[] whatsapp.Chat){
	
	//fmt.Println(contacts);
}

func (h *waHandler) HandleJsonMessage(message string) {
		
	fmt.Println("=========> HandleJsonMessage <=========")
	fmt.Println(message)

	var msg []interface{}
	err := json.Unmarshal([]byte(message), &msg)

	if err != nil {
			log.Printf("json parsing error %v", err)
			return
	}

	if msg[0] == "Stream" {
		if msg[1] == "asleep" {
			/*force reconnect*/
			fmt.Println("=========> ENTROU EM MODO REPOUSO <=========")
			//             go wac.NotifyConnectivity(false)
			// wh.SendSignal(PAUSE)
		}
	}

	if msg[0] == "Conn" {
		/*DADOS DO TELEFONE*/
		// fmt.Printf("=============== DADOS DO TELEFONE ===============")
		battery := h.c.Info.Battery
		wifi := h.c.Info.Connected
		charge := h.c.Info.Plugged
		is24h := h.c.Info.Is24h
		platform := h.c.Info.Platform

		


		responseDevice := deviceInfo{Info: msg}
		
		responseDevice.Jid = h.jid
		responseDevice.PhoneNumber = h.c.Info.Wid
		responseDevice.Battery = battery
		responseDevice.Wifi = wifi
		responseDevice.Charge = charge
		responseDevice.Is24h = is24h
		responseDevice.Platform = platform

		
		
		jsonStr, _ := json.Marshal(responseDevice)

		urlPost := hlp.Config.GetString("SERVER_API_NODE_DEVICE")
		req, _ := http.NewRequest("POST", urlPost, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			hlp.LogPrintln("error", "http-server", err.Error())
			return
		}


		defer resp.Body.Close()	
	}

}

// func (h *waHandler) HandleBatteryMessage(message whatsapp.BatteryMessage) {
// 	fmt.Println("=========> HandleBatteryMessage")
// 	fmt.Println(message)
// }


func WASyncVersion(conn *whatsapp.Conn) (string, error) {
	//conn.SetClientVersion(0, 4,  2081)
	// conn.SetClientVersion(2, 2121, 6)
	// conn.SetClientVersion(2, 2142, 12)
	// conn.SetClientVersion(2, 2140, 12)
	//conn.SetClientVersion(2, 2140, 12)
	conn.SetClientVersion(3, 3234, 9)
	versionClient := conn.GetClientVersion()
	
	return fmt.Sprintf("whatsapp version %v.%v.%v", versionClient[0], versionClient[1], versionClient[2]), nil
}

func WATestPing(conn *whatsapp.Conn) error {
	ok, err := conn.AdminTest()
	if !ok {
		if err != nil {
			return err
		} else {
			return errors.New("something when wrong while trying to ping, please check phone connectivity")
		}
	}

	return nil
}

func WAGenerateQR(timeout int, chanqr chan string, qrstr chan<- string) {
	select {
	case tmp := <-chanqr:
		png, _ := qrcode.Encode(tmp, qrcode.Medium, 256)
		qrstr <- base64.StdEncoding.EncodeToString(png)
	}
}

func WASessionInit(jid string, timeout int) error {
	if wac[jid] == nil {
		conn, err := whatsapp.NewConn(time.Duration(timeout) * time.Second)
		if err != nil {
			return err
		}
		//conn.SetClientVersion(0, 4,  2081)
		// conn.SetClientVersion(2, 2121, 6)
		//conn.SetClientVersion(2, 2142, 12)
		conn.SetClientVersion(3, 3234, 9)
		conn.SetClientName("Hiperchat","Chrome","10.15.4")


		info, err := WASyncVersion(conn)
		if err != nil {
			return err
		}
		hlp.LogPrintln(hlp.LogLevelInfo, "whatsapp", info)

		
		wac[jid] = conn

	}

	return nil
}

func WASessionLoad(file string) (whatsapp.Session, error) {
	session := whatsapp.Session{}

	buffer, err := os.Open(file)
	if err != nil {
		return session, err
	}
	defer buffer.Close()

	err = gob.NewDecoder(buffer).Decode(&session)
	if err != nil {
		return session, err
	}

	return session, nil
}

func WASessionSave(file string, session whatsapp.Session) error {
	arq, err := os.Create(file)
	if err != nil {
		return err
	}
	defer arq.Close()
	encoder := gob.NewEncoder(arq)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}

func WASessionExist(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		return false
	}

	return true
}

func WASessionConnect(jid string, timeout int, file string, qrstr chan<- string, errmsg chan<- error) {
	chanqr := make(chan string)

	session, err := WASessionLoad(file)
	if err != nil {
		go func() {
			WAGenerateQR(timeout, chanqr, qrstr)
		}()
		
		err = WASessionLogin(jid, timeout, file, chanqr)
		if err != nil {
			errmsg <- err
			return
		}
		return
	}

	err = WASessionRestore(jid, timeout, file, session)
	if err != nil {
		go func() {
			WAGenerateQR(timeout, chanqr, qrstr)
		}()

		err = WASessionLogin(jid, timeout, file, chanqr)
		if err != nil {
			errmsg <- err
			return
		}
	}

	err = WATestPing(wac[jid])
	if err != nil {
		errmsg <- err
		return
	}

	errmsg<- errors.New("")
	return
}

func WASessionLogin(jid string, timeout int, file string, qrstr chan<- string) error {
	
	if wac[jid] != nil {
			/*if WASessionExist(file) {
			err := os.Remove(file)
			if err != nil {
				return err
			}
		}
		*/
		delete(wac, jid)
	}

	err := WASessionInit(jid, timeout)
	
	if err != nil {
		return err
	}

	session, err := wac[jid].Login(qrstr)
	if err != nil {
		switch strings.ToLower(err.Error()) {
		case "already logged in":
			return nil
		case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
			notifyError(jid)
			delete(wac, jid)
			return errors.New("connection is invalid")
		default:
			notifyError(jid)
			delete(wac, jid)
			return err
		}
	}

	err = WASessionSave(file, session)
	if err != nil {
		return err
	}

	wac[jid].AddHandler(&waHandler{wac[jid], jid, uint64(time.Now().Unix())})

	return nil
}

func WASessionRestore(jid string, timeout int, file string, sess whatsapp.Session) error {
	
	if wac[jid] != nil {
	
	/*	if WASessionExist(file) {
			err := os.Remove(file)
			if err != nil {
				return err
			}
		}*/

		delete(wac, jid)
	}

	err := WASessionInit(jid, timeout)
	if err != nil {
		return err
	}

	session, err := wac[jid].RestoreWithSession(sess)
	if err != nil {
		switch strings.ToLower(err.Error()) {
		case "already logged in [Jid: "+jid+"]":
			return nil
		case "could not send proto: failed to write message: error writing to websocket: websocket: close sent [Jid: "+jid+"]":
			notifyError(jid)
			delete(wac, jid)
			return errors.New("connection is invalid [Jid: "+jid+"]")
		default:
			notifyError(jid)
			delete(wac, jid)
			return err
		}
	}

	err = WASessionSave(file, session)
	if err != nil {
		return err
	}

	wac[jid].AddHandler(&waHandler{wac[jid], jid, uint64(time.Now().Unix())})
	
	return nil
}

func WASessionLogout(jid string, file string) error {
	if wac[jid] != nil {
		err := wac[jid].Logout()
		if err != nil {
			return err
		}

		
		if WASessionExist(file) {
			err = os.Remove(file)
			if err != nil {
				return err
			}
		}

		delete(wac, jid)
	} else {
		notifyError(jid)
		return errors.New("connection is invalid [Jid: "+jid+"]")
	}

	return nil
}

func WAMessageText(jid string, jidDest string, msgText string, msgQuotedID string, msgQuoted string, msgDelay int) (string, error) {
	var id string
	if wac[jid] != nil {
		jidPrefix := "@s.whatsapp.net"
		if len(strings.SplitN(jidDest, "-", 2)) == 2 {
			jidPrefix = "@g.us"
		}

		content := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: jidDest + jidPrefix,
			},
			Text: msgText,
		}

		<-time.After(time.Duration(msgDelay) * time.Second)
		id, err := wac[jid].Send(content)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "sending message timed out: [Jid: "+jid+"]":
				notifyError(jid)
				return id, nil
			case "could not send proto: failed to write message: error writing to websocket: websocket: close sent [Jid: "+jid+"]":
				notifyError(jid)
				delete(wac, jid)
				return "", errors.New("connection is invalid [Jid: "+jid+"]")
			default:
				return "", err
			}
		}
	} else {
		notifyError(jid)
		return "", errors.New("connection is invalid [Jid: "+jid+"]")
	}

	return id, nil
}

func WAMessageImage(jid string, jidDest string, msgImageStream multipart.File, msgImageType string, msgCaption string, msgDelay int) (string, error) {
	var id string

	if wac[jid] != nil {
		jidPrefix := "@s.whatsapp.net"
		if len(strings.SplitN(jidDest, "-", 2)) == 2 {
			jidPrefix = "@g.us"
		}

		content := whatsapp.ImageMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: jidDest + jidPrefix,
			},
			Content: msgImageStream,
			Type:    msgImageType,
			Caption: msgCaption,
		}

		<-time.After(time.Duration(msgDelay) * time.Second)

		id, err := wac[jid].Send(content)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "sending message timed out [Jid: "+jid+"]":
				notifyError(jid)
				return id, nil
			case "could not send proto: failed to write message: error writing to websocket: websocket: close sent [Jid: "+jid+"]":
				delete(wac, jid)
				notifyError(jid)
				return "", errors.New("connection is invalid [Jid: "+jid+"]")
			default:
				return "", err
			}
		}
	} else {
		notifyError(jid)
		return "", errors.New("connection is invalid [Jid: "+jid+"]")
	}
	return id, nil
}

func WAMessageDocument(jid string, jidDest string, msgDocumentStream multipart.File, msgDocumentType string, msgDocumentInfo string, msgDelay int) (string, error) {
	var id string

	if wac[jid] != nil {
		jidPrefix := "@s.whatsapp.net"
		if len(strings.SplitN(jidDest, "-", 2)) == 2 {
			jidPrefix = "@g.us"
		}

		content := whatsapp.DocumentMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: jidDest + jidPrefix,
			},
			Content:  msgDocumentStream,
			Type:     msgDocumentType,
			FileName: msgDocumentInfo,
			Title:    msgDocumentInfo,
		}

		<-time.After(time.Duration(msgDelay) * time.Second)

		id, err := wac[jid].Send(content)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "sending message timed out":
				notifyError(jid)
				return id, nil
			case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
				notifyError(jid)
				delete(wac, jid)
				return "", errors.New("connection is invalid")
			default:
				return "", err
			}
		}
	} else {
		notifyError(jid)
		return "", errors.New("connection is invalid")
	}

	return id, nil
}

/*SEND AUDIO */
func WAMessageAudio(jid string, jidDest string, msgAudioStream multipart.File, msgAudioType string, msgAudioInfo string, msgDelay int) (string, error) {
	var id string
	// var idMsg string
	// audioFile, _ := os.Open("output.opus")
	// idMsg = guuid.New().String()
	b := make([]byte, 10)
	rand.Read(b)
	msgID := strings.ToUpper(hex.EncodeToString(b))
	
	if wac[jid] != nil {
		jidPrefix := "@s.whatsapp.net"
		if len(strings.SplitN(jidDest, "-", 2)) == 2 {
			jidPrefix = "@g.us"
		}

		content := whatsapp.AudioMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: jidDest + jidPrefix,
				Timestamp: uint64(time.Now().Unix()),
				Id: msgID,//"456",
			},
				Type: "audio/ogg; codecs=opus",
                Content: msgAudioStream,
                Ptt: true,
		}

		<-time.After(time.Duration(msgDelay) * time.Second)

		wac[jid].Send(content)

		// id, err := wac[jid].Send(content)
		// if err != nil {
		// 	switch strings.ToLower(err.Error()) {
		// 	case "sending audio timed out [Jid: "+jid+"]":
		// 		notifyError(jid)
		// 		fmt.Println(jid)
		// 		return id, nil
		// 	case "could not send proto: failed to write message: error writing to websocket: websocket: close sent [Jid: "+jid+"]":
		// 		delete(wac, jid)
		// 		notifyError(jid)
		// 		return "", errors.New("connection is invalid [Jid: "+jid+"]")
		// 	default:
		// 		return "", err
		// 	}
		// }
	} else {
		notifyError(jid)
		return "", errors.New("connection is invalid [Jid: "+jid+"]")
	}
	return id, nil
}

func notifyError(jid string) {
	var errorDisconnected = errorDisconnected{jid};
	jsonStr, _ := json.Marshal(errorDisconnected)
	urlPost := hlp.Config.GetString("SERVER_API_NODE_ERROR")
	req, _ := http.NewRequest("POST", urlPost, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		hlp.LogPrintln("error", "http-server", err.Error())
		return
	}

	defer resp.Body.Close()	
	}

