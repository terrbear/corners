# golang:1.16.5-alpine3.13
FROM golang@sha256:c8cbfe53dae2363fa89a7359bdf2d2e7a82239ea36fe8ea5444cc365983331e6 as builder

RUN apk add --no-cache git make 

# Configure Go
ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH /usr/local/go/bin/bin:$PATH

RUN mkdir -p ${GOPATH}/src /trustgrid

COPY . ${GOPATH}/src

WORKDIR $GOPATH/src

RUN make build_server && \
      mv bin/* /corners

# 3.13.2
FROM alpine@sha256:08d6ca16c60fe7490c03d10dc339d9fd8ea67c6466dea8d558526b1330a85930

RUN mkdir -p /corners

COPY --from=builder /corners /corners

WORKDIR /corners

EXPOSE 9007

CMD "./corners"
