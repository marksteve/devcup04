FROM golang:1.4
RUN go get github.com/tools/godep
ENV WORKDIR /go/src/github.com/marksteve/radioslack
WORKDIR $WORKDIR
RUN mkdir -p $WORKDIR
COPY . $WORKDIR
CMD ["godep", "go", "run", "main.go"]
