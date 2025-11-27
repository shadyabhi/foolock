FROM alpine:3.21
COPY foolock /foolock
ENTRYPOINT ["/foolock"]
