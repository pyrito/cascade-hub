all:
	go build main.go buffer.go controller.go device.go

clean :
	-rm main;