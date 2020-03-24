all: clean build_windows build_linux

clean:
	-rm -rf covid-19-reporter covid-19-reporter.exe

build_windows:
	GOOS=windows go build -o covid-19-reporter.exe .

build_linux:
	go build -o covid-19-reporter .