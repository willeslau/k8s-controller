# k8s-controller
An implementation of a simple k8s controller:
1. Create custom resource
2. Launch a RPC worker pod with specific memory, cpu and concurrency limit

## Generate K8S code
```bash
docker build -t code-gen .
docker run -it --rm -v `pwd`:/go/src/github.com/willeslau/k8s-controller code-gen bash

# inside docker
/go/src/k8s.io/code-generator/generate-groups.sh all \
github.com/willeslau/k8s-controller/pkg/client \
github.com/willeslau/k8s-controller/pkg/apis \
worker:v1
```
K8S generated code should be populated.

## Write the controller
The code is in `internal/controller`

## Launch the controller
Get the master location by `kubectl cluster-info`
```bash
./controller --kubeconfig ~/.kube/config --master https://192.168.1.88:6443
```