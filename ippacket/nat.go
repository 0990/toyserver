package ippacket

import (
	"errors"
	"net"
	"strconv"
	"sync"
)

const(
	MAXClient = 200
)

const(
	IPSTtart = 2
	IPEnd = IPSTtart+MAXClient
	)

type NAT struct{
	sync.Mutex
	iptable map[RawIPv4]net.Conn
	ippool map[net.Conn]RawIPv4
	allocCount int
	prefix string
}

func NewNAT(subnet string)(*NAT){
	nat:=NAT{
		Mutex:      sync.Mutex{},
		iptable:    make(map[RawIPv4]net.Conn),
		ippool:     make(map[net.Conn]RawIPv4),
		allocCount: 0,
		prefix:     subnet,
	}
	return &nat
}

func(p *NAT)Add(conn net.Conn)(string,error){
	p.Lock()
	defer p.Unlock()

	if p.allocCount>=MAXClient{
		return "",errors.New("no more ip for allocate")
	}

	alloc:=IPSTtart+p.allocCount
	p.allocCount++

	ip:=p.prefix+"."+strconv.Itoa(alloc)
	rawip:=parseIP(ip)
	p.ippool[conn] = rawip
	p.iptable[rawip] = conn
	return ip,nil
}

func(p *NAT)Del(conn net.Conn)(){
	p.Lock()
	defer p.Unlock()

	rawip,ok:=p.ippool[conn]
	if ok{
		delete(p.ippool,conn)
		delete(p.iptable,rawip)
		p.allocCount--
	}
}

func(p *NAT)GetClientRaw(rawip RawIPv4)(net.Conn){
	conn,ok:= p.iptable[rawip]
	if ok{
		return conn
	}
	return nil
}

func parseIP(s string)RawIPv4{
	var ip uint32

	addr:= net.ParseIP(s)
	p:=([]byte)(addr)

	ip = uint32(p[12])
	ip |= uint32(p[13]) << 8
	ip |= uint32(p[14]) << 16
	ip |= uint32(p[15]) << 24

	return (RawIPv4)(ip)
}
