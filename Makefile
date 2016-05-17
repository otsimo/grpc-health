.PHONY: default build clean

gcrenv="prod"
imgorg="otsimo"

default: build

build:
	sh build.sh local $(gcrenv) $(imgorg)

docker:
	sh build.sh docker $(gcrenv) $(imgorg)

fmt:
	goimports -w main.go

clean:
	rm -rf bin
