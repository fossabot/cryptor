package node

import (
	"bytes"
	"encoding/gob"
	"net"

	"cpl.li/go/cryptor/crypt"
	"cpl.li/go/cryptor/p2p"
	"cpl.li/go/cryptor/p2p/proto"
)

func (n *Node) run() {
	for {
		select {
		// pick up and display errors
		case err := <-n.comm.err:
			if err != nil {
				n.logger.Println("err", err)
			}
		// receive exit signal
		case <-n.comm.exi:
			return
		}
	}
}

func (n *Node) connect() (err error) {
	// lock network state
	n.net.Lock()
	defer n.net.Unlock()

	// network bind on entire local address space, using random port
	n.net.conn, err = net.ListenUDP(p2p.Network, n.net.addr)

	return err
}

func (n *Node) disconnect() error {
	// ignore if node is not running
	if !n.state.isRunning {
		return nil
	}

	// ignore if node is not connected
	if !n.state.isConnected {
		return nil
	}

	// set node as disconnected
	n.state.isConnected = false

	// disconnect network bind
	if err := n.net.conn.Close(); err != nil {
		// on error, set node back as connected
		n.state.isConnected = true

		return err
	}

	// send disconnect signal, stop listening
	n.comm.dis <- nil

	n.logger.Println("disconnected")

	return nil
}

// forward handles receiving and sending data to the network
func (n *Node) forward() {
	for {
		select {
		// disconnect
		case <-n.comm.dis:
			return
		// receive
		case pack := <-n.net.recv:
			go n.handlePacket(pack)
		// send
		// TODO Finish sending data
		case pack := <-n.net.send:
			_, err := n.net.conn.WriteToUDP(pack.MsgData[:], &net.UDPAddr{})
			if err != nil {
				n.comm.err <- err
			}
		}
	}
}

// listen checks the network for incoming connections, extracts the data
// and passes on valid packets only
func (n *Node) listen() {
	// incoming data buffer
	buffer := make([]byte, p2p.MaxUDPSize)
	reader := bytes.NewReader(buffer)

	// zero buffer on disconnect
	defer crypt.ZeroBytes(buffer)

	for {
		// check if still connected
		if !n.state.isConnected {
			return
		}

		// read from network
		r, addr, err := n.net.conn.ReadFromUDP(buffer)
		if err != nil {
			// if disconnected return without error
			if !n.state.isConnected {
				return
			}

			// send error to node and retry
			n.comm.err <- err
			continue
		}

		// check connection type
		if p, ok := n.lookup.address[addr.String()]; ok {
			// known peer
			n.logger.Println("peer connection from", p.staticPublicKey.ToHex())
		} else {
			// unknown
			n.logger.Println("unknown connection from", addr.String())
		}

		// check for handshake messages
		switch r {
		case proto.MsgSizeHandshakeI:
			var msg proto.MsgHandshakeI
			dec := gob.NewDecoder(reader)
			n.comm.err <- dec.Decode(&msg)
		case proto.MsgSizeHandshakeR:
		}

		// send parsed packet
		n.net.recv <- nil
	}
}

func (n *Node) handlePacket(pack *proto.Packet) {
	n.logger.Println("debug pack:", pack)
}
