package localip

import (
	"log"
	"net"
)

// GetLocalIP returns the local IP address of the machine
func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
