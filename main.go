package main

import (
	"flag"
	"fmt"
	"github.com/0990/toyserver/ippacket"
	"log"
	"github.com/songgao/water"
	"net"
	"strconv"
)

const(
	HEAD_SIZE = 2
)

var (
	subnet = flag.String("local","10.0.0","local tun interface IP prefix like 10.0.0")
	dns = flag.String("d","8.8.8.8","DNS server IP like 8.8.8.8")
	bind = flag.String("bind",":2010","bind port for communication")
	dev = flag.String("tun","tun0","tun path")
	mtu = flag.Int("m",1400,"MTU size")
	secret    = flag.String("s", "123", "Shared Secret")
)

var iface *water.Interface

func main(){
	flag.Parse()

	//nat:=ippacket.NewNAT(*subnet)

	iface1,err:=water.New(water.Config{
		DeviceType:             water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			ComponentID:   "",
			InterfaceName: *dev,
			Network:       "",
		},
	})

	if err!=nil{
		log.Fatalln("unable to allocate tun interface:",err)
	}

	iface = iface1


	lis, err := net.ListenPacket("udp", *bind)
	if err!=nil{
		log.Fatalln(err)
	}


	var udpAddr net.Addr
	data := make([]byte, 64 * 1024)
	for {
		n, addr, err := lis.ReadFrom(data)
		if err != nil {
			continue
		}
		if udpAddr==nil{
			fmt.Println("first connect",addr.String())
			handshake(lis,addr,data[:n])
			udpAddr = addr
		}else{
			if udpAddr==addr{
				//drop control frame
				if data[0]==0{
					continue
				}
				iface.Write(data[:n])
			}
		}
	}

	go func(){
		packet:=make([]byte,64 * 1024)
		for{
			n,err:=iface.Read(packet)
			if err!=nil{
				break
			}

			var header = (ippacket.Packet)(packet[:n])

			fmt.Println("-------Send %d bytes\n", n)
			fmt.Println( "Dst: %s\n", header.Dst())
			fmt.Println("Src: %s\n", header.Src())
			fmt.Println("Protocol: % x\n", header.Protocol())

			if udpAddr!=nil{
				lis.WriteTo(packet[:n],udpAddr)
			}
		}
	}()

}


func handshake(conn net.PacketConn, addr net.Addr,packet []byte) (bool) {
	if len(packet)!=len(*secret)+1{
		return false
	}

	if string(packet[1:])!=*secret{
		return false
	}

	for i:=0;i<3;i++{
		param := build_parameters("10.0.0.2")
		conn.WriteTo(param,addr)
	}
	return false
}

func build_parameters(ip string) ([]byte) {
	str := "m," + strconv.Itoa(*mtu)
	str += " d," + *dns
	str += " r,0.0.0.0,0"
	str += " a," + ip + ",32"

	buf := []byte(str)
	buflen := byte(len(buf))
	buf = append([]byte{0, buflen}, buf...)
	return buf
}

func handleWrite(conn net.PacketConn,addr net.Addr){
	packet:=make([]byte,64 * 1024)
	for{
		n,err:=iface.Read(packet)
		if err!=nil{
			break
		}

		var header = (ippacket.Packet)(packet[:n])

		fmt.Println("-------Send %d bytes\n", n)
		fmt.Println( "Dst: %s\n", header.Dst())
		fmt.Println("Src: %s\n", header.Src())
		fmt.Println("Protocol: % x\n", header.Protocol())

		conn.WriteTo(packet[:n],addr)
	}
}

