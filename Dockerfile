FROM alpine:3.3
MAINTAINER "Sercan Degirmenci <sercan@otsimo.com>"

ADD grpc-health /etc/otsimo/grpc-health

CMD ["/etc/otsimo/grpc-health"]

