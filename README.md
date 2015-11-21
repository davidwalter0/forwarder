### k8s-simple-forward

Trivial forwarder to front applications from a service to a known
accessible node.

For example with the dockerized version after the build and dockerize
scripts are executed, 8080 might be forwarded from a known DNS
reachable k8s service to the local node.

Using a nodeSelector to map to a specific node on the edge of the
network with an accessible ip.

```
#!/bin/bash
# k8s-simple-forward 
dockerize=0
kuberize=0
git clone git@github.com:davidwalter0/k8s-simple-forward.git

cd k8s-simple-forward

./build

./dockerize

# run the docker service as 
if (( dockerize )); then
  docker run --name=known-service -d --net=host k8s-simple-forwarder:latest /k8s-simple-forwarder 10.10.10.6:8080 known-service:8080
fi
# or a replica using
if (( kuberize )); then
  kubectl create -f k8s-simple-forwarder-template.yaml
fi

```
