### forwarder

```
go get github.com/davidwalter0/forwarder
```

Trivial ingress forwarder for tcp connections from external endpoints to kubernetes cluster services, fixed node applications, or internal ips to provide L3 access in the cluster.

The process for building is

```
go get github.com/davidwalter0/forwarder
make
```

deploying to the cluster requires configuring a soft link to a cluster
config so that the cluster's kubeconfig is available in a subdirectory
like
```
export kubectl=${GOPATH}/bin/kubectl --kubeconfig=${PWD}/cluster/auth/kubeconfig
```
or modify the Makefile to point to a valid cluster config

A template exists for a daemonset & secret with a configuration
file. The example uses `github.com/davidwalter0/applytmpl` to generate
the config options for a kubectl config file

Using a daemonset definition, created by `make yaml` then make apply
transforms the pipes.yaml to base64 encoding and injects it in place
in the daemonset.yaml configuration. `make apply` executes `kubectl
apply -f` for the setup for forwarding

```
make yaml
make apply
```


An example yaml config file with one or more services or direct
point to point connections from outside to inside the cluster

The formats accepted are maps of 

```
name:
  source: ip:port
  sink: ip:port
```

or

```
name:
  source: ip:port
  service: svc
  namespace: name
  enableep: true/false
```


Example format 0: using fixed source/sink ip:port pairs

```
ssh0:
  source: "0.0.0.0:2220"
  sink: "10.2.0.33:22"
```

Example format 0': Use the service.namespace as the sink

```
ssh1:
  source: "0.0.0.0:2221"
  sink: "ssh.default:22"
```

Example format 1: using cluster's service endpoints directly bypassing
kubernetes internal scheduling, but points directly to cluster endpoints

```
ssh2:
  source: "0.0.0.0:2222"
  service: ssh
  namespace: default
  enableep: true
```

Example format 2: using cluster's service with kubernetes internal scheduling to
select endpoints

```
ssh3:
  source: "0.0.0.0:2223"
  service: ssh
  namespace: default
  enableep: false
```

TODO
- [X] Add yaml daemonset config option for environment variable for default file location
- [X] Add volume mount for file
- [X] Add pipes.yaml
- [X] Add file change monitoring and reload
- [X] Add multiple endpoint select
- [ ] Unit Test Reload
- [ ] Unit Test Kill and Restart go routines
- [ ] Add service watcher for endpoint changes
- [ ] Add mgmt monitor for concurrent access/update/use of listeners
