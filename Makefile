all:
	go build src/main.go src/buffer.go src/controller.go src/device.go

clean :
	-rm main;