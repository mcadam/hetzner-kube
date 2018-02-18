FROM scratch

#RUN apt-get update && apt-get install -y openssh-client
ADD ca-certificates.crt /etc/ssl/certs/
ADD hetzner-kube /

VOLUME /sshkey

CMD ["/hetzner-kube"]
