COVID-19-Reporter
===

The main goal of the project is to monitor the state of the COVID-19 coronavirus and notify via email at the current time about the state of the virus via email.

## Content

- [Configuration](#configuration)
- [How to start](#how-to-start)
- [Run as service](#run-as-service)
    - [For Linux](#for-linux)

## Configuration

All settings are stored in the `config.env`. File structure (example for Gmail):

```
SMTP_SERVER=smtp.gmail.com
SMTP_PORT=587
EMAIL=example@gmail.com
PASSWORD=example_password
REPORT_PERIOD=60
REPORT_TO=example1@gmail.com,example2@gmail.com
```

Settings **SMTP_PORT** and **REPORT_PERIOD** are numbers, and **REPORT_PERIOD** is the number of minutes to resend the report.

Setting **REPORT_TO** may contain multiple emails, separated by commas.

If there is no settings file, an exception is thrown.

## How to start

To start, you can use the command line:

```
go run .
```

or

```
go build . && ./covid-19-reporter
```

Or build the application using the `Makefile` by calling the command:

```
make
```

this will create the files `covid-19-reporter` for Linux, and `covid-19-reporter.exe` for Windows.

## Run as service

### For Linux

In order to run the application as a service, you must:

1. Build the application with the ```make``` command.
2. In the `covid-19-reporter.service` file, specify the correct path where the `covid-19-reporter` file obtained after the assembly is located.
3. Copy the file `covid-19-reporter.service` to the directory `/etc/systemd/system/`.
4. Run the service with commands:
```
systemctl enable covid-19-reporter
systemctl start covid-19-reporter
```

At the same time, to view service messages, you must call the command:
```
systemctl -fu covid-19-reporter
```