# k8s-controller
A simple implementation of k8s controller


```bash
docker build -t code-gen .

docker run -it --rm -v `pwd`:/go/src/github.com/willeslau/k8s-controller code-gen bash
/go/src/k8s.io/code-generator/generate-groups.sh all \
github.com/willeslau/k8s-controller/pkg/client \
github.com/willeslau/k8s-controller/pkg/apis \
worker:v1
```