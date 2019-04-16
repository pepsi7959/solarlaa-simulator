all: bin

clean:
	rm -rf main solarlaa-simulator

bin:
	go build -o solarlaa-simulator main.go
