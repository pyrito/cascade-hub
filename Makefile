BIND = ./bin
OUTB = main
SRCF = $(wildcard src/*.go)

OUTPUT = $(addprefix $(BIND)/,${OUTB})

${OUTPUT}: ${BIND}/% : ${SRCF} ${BIND}
	go build -o ${BIND}/$* ${SRCF}

${BIND}:
	mkdir -p ${BIND}

run: ${BIND}/main
	${BIND}/main --devices 1

clean :
	-rm bin/*
