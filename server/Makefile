build:
	go build -o ./spock_server *.go

run: build
	./spock_server

watch:
	clear
	ulimit -n 1000
	reflex -s -r '\.go$$' make run