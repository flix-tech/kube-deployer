FROM alpine

RUN apk --no-cache add ca-certificates git

ADD kube-deploy-linux /usr/local/bin/kube-deploy
RUN chmod +x /usr/local/bin/kube-deploy

ADD https://storage.googleapis.com/kubernetes-release/release/v1.7.7/bin/linux/amd64/kubectl /usr/local/bin/kubectl
RUN chmod +x /usr/local/bin/kubectl
