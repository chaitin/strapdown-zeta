FROM alpine
ADD server/strapdown-server /strapdown-server
ENTRYPOINT ["/strapdown-server"]
