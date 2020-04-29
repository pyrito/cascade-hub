all:
	go build src/main.go src/helper.go src/controller.go src/device.go

clean :
	-rm main;
