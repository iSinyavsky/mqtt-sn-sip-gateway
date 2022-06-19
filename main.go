package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mqttsip/common"
	"net"
	"strconv"
	"strings"
	"time"
)

var asteriskIp, fromId, toId, srcIp, srcMQTTPort, brokerIp string
var asteriskPort int
var isRegister bool
var mode string

func main() {
	config := make(map[string]string)

	common.IsDebug = true
	content, err := ioutil.ReadFile("mqttsip.json")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	if err := json.Unmarshal(content, &config); err != nil {
		panic(err)
	}

	asteriskIp = config["asteriskIp"]
	fromId = config["fromId"]
	toId = config["toId"]
	srcIp = config["srcIp"]
	srcMQTTPort = config["srcMQTTPort"]
	asteriskPort, _ = strconv.Atoi(config["asteriskPort"])
	mode = config["mode"]
	brokerIp = config["brokerIp"]

	common.Print(asteriskIp, fromId, toId, srcIp, srcMQTTPort)
	listenUDPPort := asteriskPort
	if mode == "broker" {
		listenUDPPort = 5061
	}
	listener, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: listenUDPPort}) // listening udp socket
	for {
		handleClient(listener) // handle client request
	}
}

func sendSIP(conn *net.UDPConn, data []byte) {
	//protocolData := ""
	hex := common.BytesToHex(data)
	contentLength := len(hex)
	randomStr := common.GetMD5Hash(fmt.Sprintf("%d", time.Now().UnixNano()))[:5]
	sip := "MESSAGE sip:" + toId + "@" + asteriskIp + " SIP/2.0" + "\r\n" +
		"Via: SIP/2.0/UDP " + srcIp + ";branch=z9hG4bK" + randomStr + " \r\n" +
		"Max-Forwards: 70" + "\r\n" +
		"From: sip:" + fromId + "@" + asteriskIp + ";tag=" + fromId + "\r\n" +
		"To: sip:" + toId + "@" + asteriskIp + "\r\n" +
		"Call-ID: " + randomStr + "\r\n" +
		"CSeq: 1 MESSAGE" + "\r\n" +
		"Content-Type: text/plain" + "\r\n" +
		"Content-Length: " + strconv.Itoa(contentLength) + "\r\n\r\n" + hex
	conn.WriteToUDP([]byte(sip), &net.UDPAddr{IP: net.ParseIP(asteriskIp), Port: asteriskPort})
	common.Print("SIP sent", hex)
}
func sendMsgOk(conn *net.UDPConn, buff []byte) {
	msg := string(buff)
	msg = strings.Split(msg, "\r\n\r\n")[0] + "\r\n\r\n"
	lines := strings.Split(msg, "\r\n")
	lines[0] = "SIP/2.0 200 Ok"
	okMsg := strings.Join(lines, "\r\n")
	conn.WriteToUDP([]byte(okMsg), &net.UDPAddr{IP: net.ParseIP(asteriskIp), Port: asteriskPort})
}
func sendOptions(conn *net.UDPConn, lines []string) {
	lines[0] = "SIP/2.0 200 Ok"
	okMsg := strings.Join(lines, "\r\n")
	conn.WriteToUDP([]byte(okMsg), &net.UDPAddr{IP: net.ParseIP(asteriskIp), Port: asteriskPort})
}
func sendRegister(conn *net.UDPConn, msg string, isConfirm bool) {

	sip := "REGISTER sip:" + asteriskIp + " SIP/2.0" + "\r\n" +
		"Via: SIP/2.0/UDP " + srcIp + ":" + fmt.Sprintf("%d", asteriskPort) + ";branch=z9hG4bK.buqQKtltD;rport" + "\r\n" +
		"From: <sip:" + fromId + "@" + asteriskIp + ">;tag=" + fromId + "\r\n" +
		"To: sip:" + fromId + "@" + asteriskIp + "\r\n" +
		"CSeq: 21 REGISTER" + "\r\n" +
		"Call-ID: YRNAOkc8g9" + "\r\n" +
		"Max-Forwards: 70" + "\r\n" +
		"Supported: replaces, outbound, gruu" + "\r\n" +
		"Accept: application/sdp" + "\r\n" +
		"Accept: text/plain" + "\r\n" +
		"Accept: application/vnd.gsma.rcs-ft-http+xml" + "\r\n" +
		"Accept: application/octet-stream" + "\r\n" +
		"Contact: <sip:" + fromId + "@" + srcIp + ";transport=udp>;+sip.instance=\"<urn:uuid:0dc4266e-9dfb-00b6-9375-db443cc2fc08>\"" + "\r\n" +
		"Expires: 123600" + "\r\n" +
		"User-Agent: SIP_Gateway/0.0.1" + "\r\n"

	if isConfirm {
		lines := strings.Split(msg, "\r\n")
		wwwAuthLine := lines[9]
		realm := strings.Split(strings.Split(wwwAuthLine, "realm=\"")[1], "\"")[0]
		nonce := strings.Split(strings.Split(wwwAuthLine, "nonce=\"")[1], "\"")[0]
		str1 := common.GetMD5Hash(fmt.Sprintf("%s:%s:%s", fromId, realm, "abc987"))
		str2 := common.GetMD5Hash(fmt.Sprintf("REGISTER:%s", "sip:"+asteriskIp))
		response := common.GetMD5Hash(fmt.Sprintf("%s:%s:%s", str1, nonce, str2))

		sip += "Authorization:  Digest realm=\"asterisk\", nonce=\"" + nonce + "\", algorithm=MD5, username=\"" + fromId + "\",  uri=\"sip:" + asteriskIp + "\", response=\"" + response + "\"\r\n"
		isRegister = true
	}

	sip += "\r\n"

	conn.WriteToUDP([]byte(sip), &net.UDPAddr{IP: net.ParseIP(asteriskIp), Port: asteriskPort})
}

var snAddr *net.UDPAddr

func handleClient(conn *net.UDPConn) {
	if !isRegister {
		sendRegister(conn, "", false)
	}
	buf := make([]byte, 1500) // buffer for client data
	common.Print("Wait packet")
	readLen, addr, err := conn.ReadFromUDP(buf) // reading from socket
	if !addr.IP.Equal(net.ParseIP("127.0.0.1")) && !addr.IP.Equal(net.ParseIP(asteriskIp)) && !addr.IP.Equal(net.ParseIP("185.18.55.216")) {
		return
	}
	msg := string(buf[:readLen])
	if err != nil {
		fmt.Println(err)
		return
	}
	if addr.Port != asteriskPort {
		common.Print("Get from broker/client", msg)
		snAddr = addr
		sendSIP(conn, buf[:readLen])
	} else if addr.Port == asteriskPort {
		firstLine := strings.Split(msg, "\r\n")[0]
		method := strings.Split(firstLine, " ")[0]

		if method == "MESSAGE" {
			mqttPartIndex := bytes.Index(buf[:readLen], []byte("\r\n\r\n")) + 4
			mqttPart := buf[mqttPartIndex:readLen]
			finalMsg := common.HexToBytes(string(mqttPart))
			common.Print("Get from SIP", "str", string(finalMsg))
			if mode == "broker" {
				snAddr = &net.UDPAddr{IP: net.ParseIP(brokerIp), Port: 2883}
			}
			sendMsgOk(conn, buf[:readLen])
			conn.WriteToUDP(finalMsg, snAddr)
			common.Print("Sent to broker/client", string(finalMsg))
		}

		if method == "OPTIONS" {
			sendOptions(conn, strings.Split(msg, "\r\n"))
		}

		if firstLine == "SIP/2.0 401 Unauthorized" {
			sendRegister(conn, msg, true)
		}
	}

}
