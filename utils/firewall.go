package utils

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	firewallPath string = "/polycube/v1/firewall/"
)

func CreateFirewall(ip string) bool {
	resp, err := http.Post("http://"+ip+":9000"+firewallPath+"fw", "application/json", nil)
	if err != nil {
		log.Infoln("Could not create firewall:", err, resp)
		return false
	}
	return true
}
