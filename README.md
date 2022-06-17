# chainsaw-cni

![chainsaw cni logo](docs/chainsaw-cni.png)

## Chainsaw: A configuration and debugging tool for rough cuts using CNI chains.

The gist is that it allows you to tweak parameters of your network namespaces at runtime. Enables you to run `ip` commands against your containers network namespace from within a CNI chain -- by annotating a pod with the `ip` commands you'd like to run.

You add chainsaw-cni as a member of a CNI chain, then... you annotate a pod -- you use `ip` commands, and it modifies your network namespace using the `ip` command.

## Installation

Clone this repository and install the daemonset with:

```
kubectl create -f deployments/daemonset.yaml
```

## Example Usage

This example uses Multus CNI to attach a second interface. (To be expanded later!)

There's two parts:

* Create a net-attach-def (or any CNI configuration) which references the `chainsaw` plugin in a CNI chain (e.g. in a `plugins` list in a CNI configuration)
* Annotate a pod using the `ip` commands that you want executed for that pod.


In the following example pay attention to this annotation:

```
k8s.v1.cni.cncf.io/chainsaw: >
      ["ip route","  ip addr"]
```

This is a JSON list of `ip` commands to run against the pod.

The net-attach-def shows the chainsaw plugin used in a chain along with the bridge plugin.

Create a net-attach-def and pod using this yaml:

```
---
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: test-chainsaw
spec:
  config: '{
    "cniVersion": "0.4.0",
    "name": "test-chainsaw-chain",
    "plugins": [{
      "type": "bridge",
      "name": "mybridge",
      "bridge": "chainsawbr0",
      "ipam": {
        "type": "host-local",
        "subnet": "192.0.2.0/24"
      }
    }, {
      "type": "chainsaw",
      "foo": "bar"
    }]
  }'
---
apiVersion: v1
kind: Pod
metadata:
  name: chainsawtestpod
  annotations:
    k8s.v1.cni.cncf.io/networks: test-chainsaw
    k8s.v1.cni.cncf.io/chainsaw: >
      ["ip route","  ip addr"]
spec:
  containers:
  - name: chainsawtestpod
    command: ["/bin/ash", "-c", "trap : TERM INT; sleep infinity & wait"]
    image: alpine
```

Next, check what node the pod is running with:

```
kubectl get pods -o wide
```

You can then find the output from the results of the `ip` commands from the chainsaw daemonset that is running on that node, e.g.

```
kubectl get pods -n kube-system -o wide | grep -iP "status|chainsaw"
```

And looking at the logs for the daemonset pod that correlates to the node on which the pod resides, for example:

```
kubectl logs kube-chainsaw-cni-ds-kgx69 -n kube-system
```

In this case we have this output:

```
Detected commands: [route addr]
Running ip netns exec 8b95825ee825d5585e3a209aa81ae44f55b15d6eef49ff35fb4a7fd8b7577879 ip route ===============
default via 10.244.0.1 dev eth0 
10.244.0.0/24 dev eth0 proto kernel scope link src 10.244.0.121 
10.244.0.0/16 via 10.244.0.1 dev eth0 
192.0.2.0/24 dev net1 proto kernel scope link src 192.0.2.147 


Running ip netns exec 8b95825ee825d5585e3a209aa81ae44f55b15d6eef49ff35fb4a7fd8b7577879 ip addr ===============
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
3: eth0@if4147: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UP group default 
    link/ether 1e:42:76:aa:cc:c8 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.244.0.121/24 brd 10.244.0.255 scope global eth0
       valid_lft forever preferred_lft forever
5: net1@if4148: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default 
    link/ether 12:da:3f:50:1b:5b brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 192.0.2.147/24 brd 192.0.2.255 scope global net1
       valid_lft forever preferred_lft forever

```

## You can modify stuff too! Of course.

Using the same net-attach-def, but this pod definition, we can modify the route.

```
---
apiVersion: v1
kind: Pod
metadata:
  name: chainsawtestpod
  annotations:
    k8s.v1.cni.cncf.io/networks: test-chainsaw
    k8s.v1.cni.cncf.io/chainsaw: >
      ["ip route add 192.0.3.0/24 dev net1", "ip route"]
spec:
  containers:
  - name: chainsawtestpod
    command: ["/bin/ash", "-c", "trap : TERM INT; sleep infinity & wait"]
    image: alpine
```

Which results in a modified route

```
Detected commands: [route add 192.0.3.0/24 dev net1 route]
Running ip netns exec 5b3207402bce6bbcb09673dc58be07490f314a4211366888bcf4fd3dfa048378 ip route add 192.0.3.0/24 dev net1 ===============


Running ip netns exec 5b3207402bce6bbcb09673dc58be07490f314a4211366888bcf4fd3dfa048378 ip route ===============
default via 10.244.0.1 dev eth0 
10.244.0.0/24 dev eth0 proto kernel scope link src 10.244.0.155 
10.244.0.0/16 via 10.244.0.1 dev eth0 
192.0.2.0/24 dev net1 proto kernel scope link src 192.0.2.177 
192.0.3.0/24 dev net1 scope link 
```

## Disclaimers

This might start out with some considerations that you want to take seriously as an administrator. There's a non-zero probability that someone can do something nasty to your network if something was missed (trying to cinch that down, still).

It's a chainsaw after all, [use it carefully](http://www.gameoflogging.com/).

## TODO

Filter expressions: Limit to just a subset of `ip` commands.