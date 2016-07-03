FROM alpine:3.4
MAINTAINER "Sercan Degirmenci <sercan@otsimo.com>"

ADD grpc-health-linux-amd64 /opt/otsimo/grpc-health

CMD ["/opt/otsimo/grpc-health"]

