package types

import (
  "encoding/json"
  "fmt"

  cniTypes "github.com/containernetworking/cni/pkg/types"
)

// NetConf is our definition for the CNI configuration
type NetConf struct {
  cniTypes.NetConf
  Foo string `json:"foo"`
}

// LoadNetConf parses our cni configuration
func LoadNetConf(bytes []byte) (*NetConf, error) {
  n := &NetConf{}
  if err := json.Unmarshal(bytes, n); err != nil {
    return nil, fmt.Errorf("failed to load netconf: %s", err)
  }
  return n, nil
}
