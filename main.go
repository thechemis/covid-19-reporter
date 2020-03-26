package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"golang.org/x/net/proxy"
)

const projectName = "COVID-19-Reporter"
const urlData = "https://www.worldometers.info/coronavirus/"
const emailSubject = "COVID-19 Report For Now"

func main() {

	log.Println(fmt.Sprintf("%s is starting...", projectName))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	exit := make(chan error)

	var (
		config *Config
		err    error
	)

	if config, err = getConfig(); err != nil {
		panic(err)
	}

	go func() {
		<-sig
		exit <- nil
	}()

	go func() {
		for {
			var data []string
			if data, err = getData(urlData, config); err != nil {
				exit <- err
				break
			}
			log.Println("New data received", data)
			if err = sendData(data, projectName, emailSubject, config); err != nil {
				exit <- err
				break
			}
			log.Println("New data sended")
			time.Sleep(time.Duration(config.ReportPeriod) * time.Minute)
		}
	}()

	if exitErr := <-exit; exitErr != nil {
		log.Println(fmt.Sprintf("Exit with error: %v", exitErr))
	} else {
		log.Println("Exit without errors")
	}

	log.Println(fmt.Sprintf("%s is stopping...", projectName))
}

func getConfig() (config *Config, err error) {

	if err = godotenv.Load("config.env"); err != nil {
		return nil, err
	}

	configVariables := []string{
		"SMTP_SERVER",
		"SMTP_PORT",
		"EMAIL",
		"PASSWORD",
		"REPORT_PERIOD",
		"REPORT_TO",
		"FOR_COUNTRY",
		"TELEGRAM_TOKEN",
		"TELEGRAM_CHAT_ID",
		"TELEGRAM_PROXY_URL",
	}

	configVariableValues := make(map[string]string)

	for _, configVariable := range configVariables {
		value := os.Getenv(configVariable)
		if value == "" {
			return nil, fmt.Errorf("Value '%s' not found in config file", configVariable)
		}
		configVariableValues[configVariable] = value
	}

	config = &Config{}

	config.SMTPServer = configVariableValues["SMTP_SERVER"]

	var port int
	if port, err = strconv.Atoi(configVariableValues["SMTP_PORT"]); err != nil {
		return nil, err
	}
	config.SMTPPort = port

	config.Email = configVariableValues["EMAIL"]
	config.Password = configVariableValues["PASSWORD"]

	var period int
	if period, err = strconv.Atoi(configVariableValues["REPORT_PERIOD"]); err != nil {
		return nil, err
	}
	config.ReportPeriod = period

	config.ReportTo = strings.Split(configVariableValues["REPORT_TO"], ",")

	config.ForCountry = configVariableValues["FOR_COUNTRY"]

	config.TelegramToken = configVariableValues["TELEGRAM_TOKEN"]
	config.TelegramChatID = configVariableValues["TELEGRAM_CHAT_ID"]
	config.TelegramProxyURL = configVariableValues["TELEGRAM_PROXY_URL"]

	return config, nil
}

func getData(url string, config *Config) (data []string, err error) {

	var (
		response *http.Response
		htmlDoc  *goquery.Document
	)

	if response, err = http.Get(url); err != nil {
		return nil, err
	}

	body := response.Body

	defer body.Close()

	if htmlDoc, err = goquery.NewDocumentFromReader(body); err != nil {
		return nil, err
	}

	commonData := htmlDoc.Find(".maincounter-number").Map(func(index int, item *goquery.Selection) string {

		title := item.Parent().Find("h1").Text()
		value := item.Text()

		title = strings.Trim(title, "\n ")
		value = strings.Trim(value, "\n ")

		value = strings.ReplaceAll(value, ",", "")

		return fmt.Sprintf("%s %s", title, value)
	})

	if len(commonData) > 0 {

		data = append(data, "Common:")
		data = append(data, "")

		for _, commonDataItem := range commonData {
			data = append(data, fmt.Sprintf("- %s", commonDataItem))
		}
	}

	if config.ForCountry != "" {

		tableItem := htmlDoc.Find("#main_table_countries_today")

		var rowItem *goquery.Selection

		tableItem.Find("tbody").First().Find("tr").Each(func(index int, item *goquery.Selection) {
			if strings.ToUpper(item.Find("td").First().Text()) == strings.ToUpper(config.ForCountry) {
				rowItem = item
			}
		})

		if rowItem != nil {

			var countryData []string

			rowItem.Find("td").Each(func(index int, item *goquery.Selection) {
				if index > 0 {

					itemText := item.Text()
					itemText = strings.Trim(itemText, " ")

					if itemText == "" {
						itemText = "-"
					}

					countryData = append(countryData, itemText)
				}
			})

			tableItem.Find("thead th").Each(func(index int, item *goquery.Selection) {
				if index > 0 {
					countryData[index-1] = fmt.Sprintf("%s: %s", item.Text(), countryData[index-1])
				}
			})

			if len(countryData) > 0 {

				if len(data) > 0 {
					data = append(data, "")
				}

				data = append(data, fmt.Sprintf("For country - %s", config.ForCountry))
				data = append(data, "")

				for _, countryDataItem := range countryData {
					data = append(data, fmt.Sprintf("- %s", countryDataItem))
				}
			}
		}
	}

	return
}

func sendData(data []string, fromTitle, subject string, config *Config) (err error) {

	messageBody := strings.Join(data, "\n")

	for _, emailTo := range config.ReportTo {

		message := fmt.Sprintf("From: %s <%s>\nTo: %s\nSubject: %s\n\n%s", fromTitle, config.Email, emailTo, subject, messageBody)

		if err = smtp.SendMail(config.GetFullSMTPServer(), smtp.PlainAuth("", config.Email, config.Password, config.SMTPServer), config.Email, []string{emailTo}, []byte(message)); err != nil {
			return err
		}
	}

	if config.TelegramToken != "" && config.TelegramChatID != "" {

		client := &http.Client{}

		if config.TelegramProxyURL != "" {

			transport := &http.Transport{}

			var dialSocksProxy proxy.Dialer
			if dialSocksProxy, err = proxy.SOCKS5("tcp", config.TelegramProxyURL, nil, proxy.Direct); err != nil {
				panic(err)
			}
			transport.Dial = dialSocksProxy.Dial

			client.Transport = transport
		}

		sendMessageURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s",
			config.TelegramToken,
			config.TelegramChatID,
			url.PathEscape(messageBody),
		)

		var response *http.Response

		if response, err = client.Get(sendMessageURL); err != nil {

			log.Println(fmt.Sprintf("Error on send telegram message: %v", err))
			err = nil
		}

		if response != nil {

			if data, err := ioutil.ReadAll(response.Body); err == nil {
				log.Println(fmt.Sprintf("Response from telegram: %s", string(data)))
			}
			err = nil
		}
	}

	return
}
