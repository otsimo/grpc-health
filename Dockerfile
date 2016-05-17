FROM alpine:3.3
MAINTAINER "Sercan Degirmenci <sercan@otsimo.com>"

ADD grpc-health-linux-amd64 /opt/otsimo/grpc-health

CMD ["/opt/otsimo/grpc-health"]

