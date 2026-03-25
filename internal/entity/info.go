package entity

type EnvInfo struct {
	IPv4       string `json:"ipv4"`
	IPv6       string `json:"ipv6"`
	OS         string `json:"os"`
	DeviceInfo string `json:"device_info"`
}
