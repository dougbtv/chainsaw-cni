#!/bin/bash
export CNI_PATH=/opt/cni/bin/
export NETCONFPATH=/tmp/cniconfig/
mkdir -p /tmp/cniconfig

cat << EOF > /tmp/cniconfig/99-test-chainsaw.conflist
{
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
}
EOF

sudo ip netns add myplayground
sudo ip netns list | grep myplayground
echo "------------------ CNI ADD"
sudo NETCONFPATH=$(echo $NETCONFPATH) CNI_PATH=$(echo $CNI_PATH) $(which cnitool) add test-chainsaw-chain /var/run/netns/myplayground
echo "------------------ CNI DEL"
sudo NETCONFPATH=$(echo $NETCONFPATH) CNI_PATH=$(echo $CNI_PATH) $(which cnitool) del test-chainsaw-chain /var/run/netns/myplayground


sudo ip netns del myplayground


# cat << EOF > /tmp/cniconfig/99-test-chainsaw.conflist
# {
#   "cniVersion": "0.4.0",
#   "name": "test-chainsaw-chain",
#   "plugins": [{
#     "type": "bridge",
#     "name": "mybridge",
#     "bridge": "chainsawbr0",
#     "ipam": {
#       "type": "host-local",
#       "subnet": "192.0.2.0/24"
#     }
#   }]
# }
# EOF