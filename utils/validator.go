package utils

import "regexp"

const (
	macV1Regex  = "^[a-fA-F0-9]{2}(:[a-fA-F0-9]{2}){5}$"
	macV2Regex  = "^([A-Fa-f0-9]{2}[-,:]){5}[A-Fa-f0-9]{2}$"
	ipv4Regex   = `^(\b25[0-5]|\b2[0-4][0-9]|\b[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$`
	userIdRegex = `^[0-9]{8}@etv[0-9]$`
	snRegex     = `^[0-9A-Za-z]{24}$`
)

// CheckMacAddressV1 检验Mac地址
func CheckMacAddressV1(mac string) bool {
	m, err := regexp.MatchString(macV1Regex, mac)
	if err != nil {
		return false
	}
	return m
}

// CheckIPv4Address 检验Mac地址
func CheckIPv4Address(ip string) bool {
	m, err := regexp.MatchString(ipv4Regex, ip)
	if err != nil {
		return false
	}
	return m
}

// CheckUserID 检验Mac地址
func CheckUserID(id string) bool {
	m, err := regexp.MatchString(userIdRegex, id)
	if err != nil {
		return false
	}
	return m
}

// CheckSNCode 简单检查SN码格式
func CheckSNCode(sn string) bool {
	m, err := regexp.MatchString(snRegex, sn)
	if err != nil {
		return false
	}
	return m
}
