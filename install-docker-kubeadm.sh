#!/bin/bash

installPackages() {

# setup firewall
ufw allow ssh
ufw allow in on eth0 to any port 51820 # vpn on private interface
ufw allow in on wg0
ufw allow 6443 # Kubernetes API secure remote port
ufw allow 80
ufw allow 443
ufw default deny incoming
ufw --force enable
ufw status verbose

# add swap for k8s 1.9 with weave
fallocate -l 2G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' | tee -a /etc/fstab

# transport stuff
apt-get update
apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common

# docker-ce
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository \
   "deb https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
   $(lsb_release -cs) \
   stable"
# kubernetes

curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF

# prepare wireguard
add-apt-repository ppa:wireguard/wireguard -y

apt-get update
apt-get install -y docker-ce kubelet kubeadm kubectl kubernetes-cni wireguard linux-headers-$(uname -r) linux-headers-virtual

# prepare for hetzners cloud controller manager
mkdir -p /etc/systemd/system/kubelet.service.d
cat > /etc/systemd/system/kubelet.service.d/90-kubelet-extras.conf << EOM
[Service]
Environment="KUBELET_EXTRA_ARGS=--fail-swap-on=false"
EOM

# prepare for docker
mkdir -p /etc/systemd/system/docker.service.d
cat > /etc/systemd/system/docker.service.d/10-docker-opts.conf << EOM
[Service]
MountFlags=shared
Environment="DOCKER_OPTS=--iptables=false --ip-masq=false"
EOM

systemctl daemon-reload
}

S=$(type -p kubeadm > /dev/null &> /dev/null; echo $?)
while [ ${S} = 1 ]; do
    echo "installing packages..."
    installPackages
    S=$(type -p kubeadm > /dev/null &> /dev/null; echo $?)
done;
