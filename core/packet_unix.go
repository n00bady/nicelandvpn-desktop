//go:build freebsd || linux || openbsd || darwin

package core

import (
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/crypto/chacha20poly1305"
)

var PREVDNS net.IP

func ReadFromLocalTunnel(MONITOR chan int) {
	defer func() {
		RecoverAndLogToFile()
		if !GLOBAL_STATE.Exiting {
			MONITOR <- 4
		} else {
			CreateLog("general", "tunnel interface loop has exited")
		}
	}()

	var (
		err            error
		waitFortimeout = time.Now()
		packetLength   int
		packetVersion  byte

		// fullData         []byte
		packet           = make([]byte, 65600)
		serializeOptions = gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
		buffer           gopacket.SerializeBuffer
		applicationLayer gopacket.ApplicationLayer
		parsedPacket     gopacket.Packet
		parsedIPLayer    *layers.IPv4
		parsedTCPLayer   *layers.TCP
		parsedUDPLayer   *layers.UDP
		// parsedDNSLayer    *layers.DNS
		// DNSQuestionDomain string
		// DomainIsBlocked   bool
		isDNSLayer bool = false

		destinationIP = [4]byte{}
		outgoingPort  *RemotePort

		encryptedPacket []byte
		lengthBytes     = make([]byte, 2)
		nonce           = make([]byte, chacha20poly1305.NonceSizeX)
		writeError      error
		writtenBytes    int
	)

WAITFORDEVICE:
	if !GLOBAL_STATE.TunnelInitialized {
		if time.Since(waitFortimeout).Seconds() > 120 {
			CreateLog("", "Adapter reader not initialized, waiting for connection")
			waitFortimeout = time.Now()
		}
		time.Sleep(500 * time.Millisecond)
		goto WAITFORDEVICE
	}

	for {

		packetLength, err = A.Interface.Read(packet)
		if err != nil {
			BUFFER_ERROR = true
			CreateLog("general", err, "error in interface reader loop")
			return
		}

		if packetLength == 0 {
			CreateLog("", "Read size was 0")
			continue
		}

		if AS == nil || !GLOBAL_STATE.Connected {
			if time.Since(waitFortimeout).Seconds() > 120 {
				CreateLog("", "ADAPTER: received packet while disconnected. This is most likely a probe packet")
				waitFortimeout = time.Now()
			}
			continue
		}

		EGRESS_PACKETS++

		packetVersion = packet[0] >> 4
		if packetVersion != 4 {
			continue
		}

		destinationIP[0] = packet[16]
		destinationIP[1] = packet[17]
		destinationIP[2] = packet[18]
		destinationIP[3] = packet[19]

		if packet[9] == 6 {
			parsedPacket = gopacket.NewPacket(packet[:packetLength], layers.LayerTypeIPv4, gopacket.Default)
			parsedIPLayer = parsedPacket.NetworkLayer().(*layers.IPv4)
			applicationLayer = parsedPacket.ApplicationLayer()
			parsedTCPLayer = parsedPacket.TransportLayer().(*layers.TCP)
			if parsedTCPLayer.RST {
				continue
			}

			outgoingPort = GetOutgoingTCPMapping(destinationIP, uint16(parsedTCPLayer.SrcPort), uint16(parsedTCPLayer.DstPort))

			if outgoingPort == nil {
				outgoingPort = CreateTCPMapping(destinationIP, uint16(parsedTCPLayer.SrcPort), uint16(parsedTCPLayer.DstPort))
				if outgoingPort == nil {
					continue
				}
			}

			parsedTCPLayer.SrcPort = layers.TCPPort(outgoingPort.Mapped)

			AS.TCPHeader.DstIP = parsedIPLayer.DstIP
			parsedIPLayer.SrcIP = AS.TCPHeader.SrcIP
			parsedTCPLayer.SetNetworkLayerForChecksum(&AS.TCPHeader)

			buffer = gopacket.NewSerializeBuffer()
			if applicationLayer != nil {
				gopacket.SerializeLayers(buffer, serializeOptions, parsedIPLayer, parsedTCPLayer, gopacket.Payload(applicationLayer.LayerContents()))
			} else {
				gopacket.SerializeLayers(buffer, serializeOptions, parsedIPLayer, parsedTCPLayer)
			}

		} else if packet[9] == 17 {
			parsedPacket = gopacket.NewPacket(packet[:packetLength], layers.LayerTypeIPv4, gopacket.Default)
			parsedIPLayer = parsedPacket.NetworkLayer().(*layers.IPv4)
			applicationLayer = parsedPacket.ApplicationLayer()
			parsedUDPLayer = parsedPacket.TransportLayer().(*layers.UDP)

			_, isDNSLayer = applicationLayer.(*layers.DNS)
			if isDNSLayer {
				PREVDNS = parsedIPLayer.DstIP
				parsedIPLayer.DstIP = C.DNSIP
				// log.Println(parsedDNSLayer)
				// if len(parsedDNSLayer.Questions) > 0 {
				// 	DNSQuestionDomain = string(parsedDNSLayer.Questions[0].Name)
				// 	// log.Println("Searching in blocklist: ", DNSQuestionDomain)
				// 	_, DomainIsBlocked = BlockedDomainMap[DNSQuestionDomain]
				// 	if DomainIsBlocked {
				// 		log.Println("IS BLOCKED: ", DNSQuestionDomain)
				// 		// DomainIsBlocked = false
				// 		continue
				// 	}
				// }
			}

			outgoingPort = GetOutgoingUDPMapping(destinationIP, uint16(parsedUDPLayer.SrcPort), uint16(parsedUDPLayer.DstPort))

			if outgoingPort == nil {
				outgoingPort = GetOrCreateUDPMapping(destinationIP, uint16(parsedUDPLayer.SrcPort), uint16(parsedUDPLayer.DstPort))
				if outgoingPort == nil {
					continue
				}
			}

			parsedUDPLayer.SrcPort = layers.UDPPort(outgoingPort.Mapped)
			AS.UDPHeader.DstIP = parsedIPLayer.DstIP
			parsedIPLayer.SrcIP = AS.UDPHeader.SrcIP
			parsedUDPLayer.SetNetworkLayerForChecksum(&AS.UDPHeader)

			buffer = gopacket.NewSerializeBuffer()
			if applicationLayer != nil {
				gopacket.SerializeLayers(buffer, serializeOptions, parsedIPLayer, parsedUDPLayer, gopacket.Payload(applicationLayer.LayerContents()))
			} else {
				gopacket.SerializeLayers(buffer, serializeOptions, parsedIPLayer, parsedUDPLayer)
			}

		} else {
			continue
		}

		if AS.TCPTunnelSocket != nil {

			encryptedPacket = AS.AEAD.Seal(nil, nonce, buffer.Bytes(), nil)
			// binary.BigEndian.PutUint16(AS.RoutingBuffer[META_DL_START:META_DL_END], uint16(len(encryptedPacket)))
			// fullData = append(CopySlice(AS.RoutingBuffer[:]), encryptedPacket...)

			binary.BigEndian.PutUint16(lengthBytes, uint16(len(encryptedPacket)))

			writtenBytes, writeError = AS.TCPTunnelSocket.Write(append(lengthBytes, encryptedPacket...))
			// writtenBytes, writeError = AS.TCPTunnelSocket.Write(fullData)
			if writeError != nil {
				BUFFER_ERROR = true
				_ = AS.TCPTunnelSocket.Close()
				return
			}

			CURRENT_UBBS += writtenBytes
			lengthBytes = make([]byte, 2)
		} else {
			GLOBAL_STATE.Connected = false
		}

		// fullData = nil

	}
}

func ReadFromRouterSocket(MONITOR chan int) {
	defer func() {
		RecoverAndLogToFile()
		if !GLOBAL_STATE.Exiting {
			CreateErrorLog("", "Router tunnel listener exiting")
			MONITOR <- 2
		}
	}()

WAIT_FOR_TUNNEL:
	if GLOBAL_STATE.ActiveRouter == nil {
		time.Sleep(500 * time.Millisecond)
		goto WAIT_FOR_TUNNEL
	}

	if AS.TCPTunnelSocket == nil {
		time.Sleep(500 * time.Millisecond)
		goto WAIT_FOR_TUNNEL
	}

	var (
		writeErr     error
		readErr      error
		writtenBytes int
		// MIDL         int = MIDBufferLength
		lengthBytes = make([]byte, 2)
		DL          uint16
		readBytes   int

		tunnelBuffer = CreateTunnelBuffer()
		nonce        = make([]byte, chacha20poly1305.NonceSizeX)
		ip           = new(layers.IPv4)
		encErr       error

		isDNSLayer       bool
		packet           []byte
		ingressPacket    gopacket.Packet
		buffer           gopacket.SerializeBuffer
		serializeOptions = gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
		appLayer         gopacket.ApplicationLayer
		TCPLayer         *layers.TCP
		UDPLayer         *layers.UDP
		incomingPort     *RemotePort
		sourceIP         = [4]byte{}
	)

	ip.TTL = 120
	ip.DstIP = TUNNEL_ADAPTER_ADDRESS_IP
	ip.Version = 4
	AS.TCPTunnelSocket.SetReadDeadline(time.Time{})

	for {

		readBytes, readErr = io.ReadAtLeast(AS.TCPTunnelSocket, lengthBytes[:2], 2)
		if readErr != nil {
			if !IGNORE_NEXT_BUFFER_ERROR {
				CreateErrorLog("", "Read: ", readErr)
				BUFFER_ERROR = true
			} else {
				IGNORE_NEXT_BUFFER_ERROR = false
			}
			return
		}

		if readBytes != 2 {
			CreateErrorLog("", "TUNNEL SMALL READ ERROR: ", AS.TCPTunnelSocket.RemoteAddr())
			return
		}

		INGRESS_PACKETS++
		DL = binary.BigEndian.Uint16(lengthBytes[0:2])

		if DL == CODE_CLIENT_new_ping {
			GLOBAL_STATE.PingReceivedFromRouter = time.Now()
			continue
		}

		_, readErr = io.ReadAtLeast(AS.TCPTunnelSocket, tunnelBuffer[:DL], int(DL))
		if readErr != nil {
			CreateErrorLog("", "TUNNEL DATA READ ERROR: ", readErr)
			return
		}

		// packet = tunnelBuffer[MIDL : MIDL+DL]
		packet, encErr = AS.AEAD.Open(nil, nonce, tunnelBuffer[:DL], nil)
		if encErr != nil {
			CreateErrorLog("", "Encryption: ", encErr)
			continue
		}

		sourceIP[0] = packet[12]
		sourceIP[1] = packet[13]
		sourceIP[2] = packet[14]
		sourceIP[3] = packet[15]
		ip.SrcIP = net.IP{sourceIP[0], sourceIP[1], sourceIP[2], sourceIP[3]}

		ingressPacket = gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.Default)
		buffer = gopacket.NewSerializeBuffer()
		appLayer = ingressPacket.ApplicationLayer()

		if packet[9] == 6 {
			ip.Protocol = 6
			TCPLayer = ingressPacket.TransportLayer().(*layers.TCP)

			incomingPort = GetTCPMapping(sourceIP, uint16(TCPLayer.DstPort))
			if incomingPort == nil {
				continue
			}

			TCPLayer.DstPort = layers.TCPPort(incomingPort.Local)

			TCPLayer.SetNetworkLayerForChecksum(ip)

			if appLayer != nil {
				gopacket.SerializeLayers(buffer, serializeOptions, ip, TCPLayer, gopacket.Payload(appLayer.LayerContents()))

			} else {
				gopacket.SerializeLayers(buffer, serializeOptions, ip, TCPLayer)
			}

		} else if packet[9] == 17 {
			ip.Protocol = 17
			UDPLayer = ingressPacket.TransportLayer().(*layers.UDP)

			incomingPort = GetUDPMapping(sourceIP, uint16(UDPLayer.DstPort))
			if incomingPort == nil {
				continue
			}

			_, isDNSLayer = appLayer.(*layers.DNS)
			if isDNSLayer {
				ip.SrcIP = PREVDNS
			}

			UDPLayer.DstPort = layers.UDPPort(incomingPort.Local)

			UDPLayer.SetNetworkLayerForChecksum(ip)

			if appLayer != nil {
				gopacket.SerializeLayers(buffer, serializeOptions, ip, UDPLayer, gopacket.Payload(appLayer.LayerContents()))
			} else {
				gopacket.SerializeLayers(buffer, serializeOptions, ip, UDPLayer)
			}

		}

		writtenBytes, writeErr = A.Interface.Write(buffer.Bytes())
		if writeErr != nil {
			CreateErrorLog("", "Send: ", writeErr)
		}
		CURRENT_DBBS += writtenBytes

		packet = nil
		buffer = nil

	}
}