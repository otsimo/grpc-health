.PHONY: default build clean

env="prod"
imgorg="otsimo"

default: build

build:
	sh build.sh local $(env) $(imgorg)

docker:
	sh build.sh docker $(env) $(imgorg)

fmt:
	goimports -w main.go

clean:
	rm -rf bin
