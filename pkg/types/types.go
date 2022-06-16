package types

import (
  "encoding/json"
  "fmt"

  cniTypes "github.com/containernetworking/cni/pkg/types"
  current "github.com/containernetworking/cni/pkg/types/040"
  cniVersion "github.com/containernetworking/cni/pkg/version"
)

// NetConf is our definition for the CNI configuration
type NetConf struct {
  cniTypes.NetConf
  PrevResult *current.Result `json:"-"`
  Foo        string          `json:"foo"`
}

// LoadNetConf parses our cni configuration
func LoadNetConf(bytes []byte) (*NetConf, error) {
  conf := NetConf{}
  if err := json.Unmarshal(bytes, &conf); err != nil {
    return nil, fmt.Errorf("failed to load netconf: %s", err)
  }

  // Parse previous result
  if conf.RawPrevResult != nil {
    resultBytes, err := json.Marshal(conf.RawPrevResult)
    if err != nil {
      return nil, fmt.Errorf("could not serialize prevResult: %v", err)
    }

    res, err := cniVersion.NewResult(conf.CNIVersion, resultBytes)

    if err != nil {
      return nil, fmt.Errorf("could not parse prevResult: %v", err)
    }

    conf.RawPrevResult = nil
    conf.PrevResult, err = current.NewResultFromResult(res)
    if err != nil {
      return nil, fmt.Errorf("could not convert result to current version: %v", err)
    }
  }

  return &conf, nil
}
