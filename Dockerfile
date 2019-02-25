FROM golang:1.11

RUN curl -qL https://github.com/Masterminds/glide/releases/download/v0.13.2/glide-v0.13.2-linux-amd64.tar.gz | tar xz

ADD . /go/src/github.com/flix-tech/kube-deployer

WORKDIR /go/src/github.com/flix-tech/kube-deployer

RUN /go/linux-amd64/glide install

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -extldflags '-static'" -o kube-deploy .

FROM alpine:3.8

RUN apk --no-cache add ca-certificates git

COPY --from=0 /go/src/github.com/flix-tech/kube-deployer/kube-deploy /usr/local/bin/kube-deploy
ADD https://storage.googleapis.com/kubernetes-release/release/v1.10.8/bin/linux/amd64/kubectl /usr/local/bin/kubectl

RUN chmod +x /usr/local/bin/kube-deploy /usr/local/bin/kubectl
