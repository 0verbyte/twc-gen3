package twc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/Ullaakut/nmap/v3"

	log "github.com/sirupsen/logrus"
)

// TWC structure that represents the TWC
type TWC struct {
	ip string
}

// IP returns the IP of the TWC
func (twc *TWC) IP() string {
	return twc.ip
}

// GetVitals returns the vitals from the TWC
func (twc *TWC) GetVitals() (*Vitals, error) {
	return getVitals(twc.ip)
}

// GetWifiStatus returns the wifi status of the TWC
func (twc *TWC) GetWifiStatus() (*WifiStatus, error) {
	return getWifiStatus(twc.ip)
}

// GetLifetime returns the lifetime stats for the TWC
func (twc *TWC) GetLifetimeStats() (*LifetimeStats, error) {
	return getLifetimeStats(twc.ip)
}

// New creates a new TWC using the given IP
func New(ip string) (*TWC, error) {
	if i := net.ParseIP(ip); i == nil {
		return nil, fmt.Errorf("%s is invalid IP", ip)
	}
	return &TWC{
		ip,
	}, nil
}

// Find attempts to locate the TWC on the local network and if found will return a TWC instance.
func Find() (*TWC, error) {
	cidr, err := getLocalCIDR()
	if err != nil {
		return nil, err
	}

	log.Debugf("Scanning network for Tesla Wall Connector (gen 3) %s", cidr.String())

	scanner, err := nmap.NewScanner(
		context.Background(),
		nmap.WithTargets(cidr.String()),
		nmap.WithPorts("80"),
		nmap.WithTimingTemplate(nmap.TimingFastest),
		nmap.WithFilterHost(func(h nmap.Host) bool {
			// Filter out hosts with no open ports.
			for idx := range h.Ports {
				if h.Ports[idx].Status() == "open" {
					return true
				}
			}

			return false
		}),
	)

	if err != nil {
		return nil, err
	}

	log.Debug("Scanning...")

	result, warnings, err := scanner.Run()
	if err != nil {
		return nil, err
	}

	if warnings != nil {
		for _, warning := range *warnings {
			log.Warnln("Scan warning", warning)
		}
	}

	for _, host := range result.Hosts {
		ip := host.Addresses[0]
		if _, err := getVitals(ip.Addr); err == nil {
			log.Debugf("Located Tesla Wall Connector at %s\n", ip.Addr)
			return &TWC{
				ip: ip.Addr,
			}, nil
		}
	}

	return &TWC{}, fmt.Errorf("unable to find TWC on the scanned network range %s", cidr.String())
}

func getLocalCIDR() (*net.IPNet, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			ip, ipNet, err := net.ParseCIDR(addr.String())
			if err != nil {
				return nil, err
			}

			if ip.IsPrivate() {
				return ipNet, nil
			}
		}
	}

	return nil, errors.New("unable to locate local private IP")
}

func getVitals(ip string) (*Vitals, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/api/1/vitals", ip))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK || resp.Header.Get("content-type") != "application/json" {
		return nil, errors.New("invalid response from API")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var vitals *Vitals
	if err := json.Unmarshal(data, &vitals); err != nil {
		return nil, err
	}

	return vitals, nil
}

func getWifiStatus(ip string) (*WifiStatus, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/api/1/wifi_status", ip))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK || resp.Header.Get("content-type") != "application/json" {
		return nil, errors.New("invalid response from API")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wifiStatus *WifiStatus
	if err := json.Unmarshal(data, &wifiStatus); err != nil {
		return nil, err
	}
	return wifiStatus, nil
}

func getLifetimeStats(ip string) (*LifetimeStats, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/api/1/lifetime", ip))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK || resp.Header.Get("content-type") != "application/json" {
		return nil, errors.New("invalid response from API")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var lifetimeStats *LifetimeStats
	if err := json.Unmarshal(data, &lifetimeStats); err != nil {
		return nil, err
	}
	return lifetimeStats, nil
}
