# k8s-controller
A simple implementation of k8s controller


```bash
docker build -t code-gen .

docker run -it --rm -v `pwd`:/go/src/k8s-controller code-gen bash
/go/src/k8s.io/code-generator/generate-groups.sh all \
k8s-controller/pkg/client \
k8s-controller/pkg/apis \
worker:v1
```