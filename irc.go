package main

import (
	"hash/fnv"
	"log"
	"net"
	"strconv"
	"strings"
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
	cWhite = iota
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

var colorMatch = map[int]int{
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
	return setIrcMode(ircColor) + strconv.Itoa(fgColor) + "," + strconv.Itoa(bgColor)
}

func colorMsg(msg string, fgColor int, bgColor int) string {
	return setIrcColor(fgColor, bgColor) + msg + setIrcMode(ircCReset)
}

func hashedColor(msg string) string {
	h := fnv.New32a()
	h.Write([]byte(msg))
	bgColor := int(h.Sum32() % 16)
	return colorMsg(msg, colorMatch[bgColor], bgColor)
}

func pad(msg string, length int) string {
    if (len(msg) > length) {
        return msg
    } else {
        return msg + strings.Repeat(" ", length-len(msg))
    }
}
