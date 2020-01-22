FROM golang:1.12
RUN go get -u github.com/golang/dep/cmd/dep && go get -u github.com/oxequa/realize