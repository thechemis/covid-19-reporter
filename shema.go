package main

import "fmt"

// Config - struct of config file
type Config struct {
	SMTPServer       string
	SMTPPort         int
	Email            string
	Password         string
	ReportPeriod     int
	ReportTo         []string
	ForCountry       string
	TelegramToken    string
	TelegramChatID   string
	TelegramProxyURL string
}

// GetFullSMTPServer - get SMTPServer:SMTPPort
func (config Config) GetFullSMTPServer() string {
	return fmt.Sprintf("%s:%d", config.SMTPServer, config.SMTPPort)
}
