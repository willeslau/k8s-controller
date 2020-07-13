# How to setup centos VM:
Virtual box
// guest addition
```bash
sudo yum install dkms binutils gcc make patch libgomp glibc-headers glibc-devel kernel-headers
sudo yum install kernel-devel
```

# Setup static ip:
```bash
# on the host
## get the gateway ip, for mac
netstat -nr | grep default

# on the guest
## get the ip
ifconfig

## edit files
nano /etc/sysconfig/network
// GATEWAY=<replace>
// NETWORKING=yes

nano /etc/sysconfig/network-scripts/ifcfg-enp0s3
// BOOTPROTO="static"
// IPADDR=<replace>
// NETMASK="255.255.255.0"
// DNS1="8.8.8.8"

systemctl restart network

// check /etc/resolv.conf has nameserver 8.8.8.8
```

# How to setup ssh to the VM:
If the network mode is bridge, then it is enabled by default.

# Setup Single node k8s:
```bash
sudo su
setenforce 0
sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config

# Stop and disable firewalld
systemctl disable firewalld --now

# Refer to k8s official doc
cat <<EOF > /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sysctl --system

modprobe br_netfilter
lsmod | grep br_netfilter

# Do not use SWAP
swapoff -a
echo "vm.swappiness = 0">> /etc/sysctl.conf
sysctl -p
nano /etc/fstab # comment off '/dev/mapper/centos-swap swap ...'
free -h

# Install docker, refer to https://docs.docker.com/engine/install/centos/
yum install -y yum-utils
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum install docker-ce docker-ce-cli containerd.io
systemctl enable docker --now
systemctl status docker

# Install k8s
cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF
yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes
systemctl enable --now kubelet
kubeadm init --pod-network-cidr=10.100.0.0/16 --service-cidr=10.101.0.0/16 --kubernetes-version=v1.18.1 --apiserver-advertise-address IPADDR

kubectl taint nodes --all node-role.kubernetes.io/master-
```

# Install IDE
As vscode is free, so we use vscode
```bash
sudo rpm --import https://packages.microsoft.com/keys/microsoft.asc
sudo sh -c 'echo -e "[code]\nname=Visual Studio Code\nbaseurl=https://packages.microsoft.com/yumrepos/vscode\nenabled=1\ngpgcheck=1\ngpgkey=https://packages.microsoft.com/keys/microsoft.asc" > /etc/yum.repos.d/vscode.repo'
yum check-update
sudo yum install code
```

# Install Golang
Follow this https://linuxize.com/post/how-to-install-go-on-centos-7/