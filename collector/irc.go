package main

import (
	"hash/fnv"
	"log"
	"net"
	"strconv"
	"strings"
)

type IrcMode int

const (
	ircBold      IrcMode = 2
	ircColor     IrcMode = 3
	ircItalic    IrcMode = 19
	ircCReset    IrcMode = 15
	ircCReverse  IrcMode = 22
	ircUnderline IrcMode = 31
)

type Color int

const (
	cWhite Color = iota
	cBlack
	cBlue
	cGreen
	cRed
	cBrown
	cPurple
	cOrange
	cYellow
	cLime
	cTeal
	cCyan
	cRoyal
	cPink
	cGrey
	cSilver
)

var colorMatch = map[Color]Color{
	cWhite:  cBlack,
	cBlack:  cWhite,
	cBlue:   cWhite,
	cGreen:  cBlack,
	cRed:    cWhite,
	cBrown:  cWhite,
	cPurple: cWhite,
	cOrange: cBlack,
	cYellow: cBlack,
	cLime:   cBlack,
	cTeal:   cBlack,
	cCyan:   cBlack,
	cRoyal:  cWhite,
	cPink:   cBlack,
	cGrey:   cBlack,
	cSilver: cBlack,
}

// IRC Bot Helper
func WriteToIrcBot(message string, conf Configuration) {
	strEcho := conf.IRCChannel + "|" + message + "\n"
	servAddr := conf.Botaddress + ":" + conf.Botport
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)

	if err != nil {
		log.Println("ResolveTCPAddr failed:", err.Error())
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)

	if err != nil {
		log.Println("Dial failed:", err.Error())
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(strEcho))

	if err != nil {
		log.Println("Write to server failed:", err.Error())
		return
	}
}

func setIrcMode(mode IrcMode) string {
	return string(byte(mode))
}

func setIrcColor(fgColor Color, bgColor Color) string {
	return setIrcMode(ircColor) + strconv.Itoa(int(fgColor)) + "," + strconv.Itoa(int(bgColor))
}

func colorMsg(msg string, fgColor Color, bgColor Color) string {
	return setIrcColor(fgColor, bgColor) + msg + setIrcMode(ircCReset)
}

func colorMatchedMsg(msg string, bgColor Color) string {
	return colorMsg(msg, colorMatch[bgColor], bgColor)
}

func hashedColor(msg string, hash string) string {
	h := fnv.New32a()
	h.Write([]byte(hash))
	bgColor := Color(h.Sum32() % 16)
	return colorMsg(msg, colorMatch[bgColor], bgColor)
}

func pad(msg string, length int) string {
	if len(msg) > length {
		return msg
	} else {
		return msg + strings.Repeat(" ", length-len(msg))
	}
}