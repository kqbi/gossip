package transport

import (
	"fmt"
	"github.com/kqbi/gossip/base"
	"github.com/kqbi/gossip/log"
	"github.com/kqbi/gossip/parser"
	"strconv"
)

import (
	"net"
)

type Udp struct {
	connTable
	listeningPoint *net.UDPConn
	output          chan base.SipMessage
	stop            bool
}

func NewUdp(output chan base.SipMessage) (*Udp, error) {
	udp := Udp{listeningPoint: nil, output: output}
	udp.connTable.Init()
	return &udp, nil
}

func (udp *Udp) Listen(address string) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	lp, err := net.ListenUDP("udp", addr)

	if err == nil {
		udp.listeningPoint =  lp
		go udp.listen(lp)
	}

	return err
}

func (udp *Udp) IsStreamed() bool {
	return false
}

func (udp *Udp) Send(addr string, msg base.SipMessage) error {
	log.Debug("Sending message %s to %s", msg.Short(), addr)
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	//var conn *net.UDPConn
	//conn, err = net.DialUDP("udp", nil, raddr)
	//if err != nil {
	//	return err
	//}
	//defer conn.Close()

	_, err = udp.listeningPoint.WriteToUDP([]byte(msg.String()),raddr)

	return err
}

func (udp *Udp) listen(conn *net.UDPConn) {
	log.Info("Begin listening for UDP on address %s", conn.LocalAddr())

	buffer := make([]byte, c_BUFSIZE)
	for {
		num, udpAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if udp.stop {
				log.Info("Stopped listening for UDP on %s", conn.LocalAddr)
				break
			} else {
				log.Severe("Failed to read from UDP buffer: " + err.Error())
				continue
			}
		}
		//addr := udpAddr.IP.String() + strconv.Itoa(udpAddr.Port)
		//received=192.0.2.1;rport=9988
		pkt := append([]byte(nil), buffer[:num]...)
		fmt.Println("pkt:", string(pkt))
		go func() {
			msg, err := parser.ParseMessage(pkt, udpAddr.IP.String(), strconv.Itoa(udpAddr.Port))
			if err != nil {
				log.Warn("Failed to parse SIP message: %s", err.Error())
			} else {
				//msg  = msg + addr;
				udp.output <- msg
			}
		}()
	}
}

func (udp *Udp) Stop() {
	udp.connTable.Stop()
	udp.stop = true
	//for _, lp := range udp.listeningPoints {
	udp.listeningPoint.Close()
	//}
}
