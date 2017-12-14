FROM alpine
ADD server/strapdown-server /strapdown-server
RUN apk update && apk add ca-certificates
ENTRYPOINT ["/strapdown-server"]
