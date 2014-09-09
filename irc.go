package main

import (
	"log"
	"net"
)

const (
	ircBold      = 2
	ircColor     = 3
	ircItalic    = 19
	ircCReset    = 15
	ircCReverse  = 22
	ircUnderline = 31
)

const (
	cWhite = '0' + iota
	cBlack
	cBlue
	cGreen
	cRed
	cBrown
	cOrange
	cYellow
	cLime
	cTeal
	cRoyal
	cPink
	cGrey
	cSilver
)

// IRC Bot Helper
func WriteToIrcBot(message string, conf Configuration) {
	strEcho := message + "\n"
	servAddr := conf.Botaddress + ":" + conf.Botport
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)

	if err != nil {
		log.Println("ResolveTCPAddr failed:", err.Error())
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	defer conn.Close()

	if err != nil {
		log.Println("Dial failed:", err.Error())
		return
	}

	_, err = conn.Write([]byte(strEcho))

	if err != nil {
		log.Println("Write to server failed:", err.Error())
		return
	}
}

func setIrcMode(mode int) string {
	return string(byte(mode))
}

func setIrcColor(fgColor int, bgColor int) string {
	var colorBytes []byte
	colorBytes = append(colorBytes, byte(ircColor), byte(fgColor), byte(','), byte(bgColor))
	return string(colorBytes)
}
