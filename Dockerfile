FROM golang:latest
COPY foolock /foolock
ENTRYPOINT ["/foolock"]
