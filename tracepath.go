/*
	Requires sudo
*/

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func main() {
	//Checking for hostname in cmd line arguments.
	if len(os.Args) < 2 {
		fmt.Println("Missing hostname or ip address")
	} else {
		hostname := os.Args[1]
		tracepath(hostname)
		fmt.Println()
	}
}

func tracepath(hostname string) {
	//creating connection
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	//Resolve IP address
	ipaddr, err := net.ResolveIPAddr("ip4:icmp", hostname)
	if err != nil {
		log.Fatalln(err.Error())
	}
	var i uint8

	//Construction ICMP message
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid(),
			Seq:  1,
			Data: []byte(""),
		},
	}

	//Marshalling ICMP message
	bin, err := msg.Marshal(nil)
	if err != nil {
		panic(err)
	}

	//Byte array to store ICMP reply
	reply := make([]byte, 1500)

	//Iterating until Max TTL or until destination
	for i = 1; i <= 255; i++ {

		//Setting TTL from zero
		conn.IPv4PacketConn().SetTTL(int(i))

		start := time.Now()
		ret, err := conn.WriteTo(bin, ipaddr)
		if err != nil {
			panic(err)
		}
		//Setting deadline for ICMP reply to 10 seconds
		err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			panic(err)
		}
		ret, peer, err := conn.ReadFrom(reply)
		//no reply from the hop. Skip and continue.
		if err != nil {
			fmt.Println(i, ": no reply")
			continue
		}
		duration := time.Since(start)

		//Parsing ICMP reply
		rm, err := icmp.ParseMessage(1, reply[:ret])

		if rm.Type == ipv4.ICMPTypeTimeExceeded {
			//TTL reached zero.
			fmt.Println(i, ":", peer, " ", duration)
		} else if rm.Type == ipv4.ICMPTypeEchoReply {
			//Destination reached
			fmt.Println(i, ":", peer, " ", duration)
			break
		}
	}
}
