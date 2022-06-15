package types

import (
	"encoding/json"
	"fmt"

	cniTypes "github.com/containernetworking/cni/pkg/types"
)

type NetConf struct {
	cniTypes.NetConf
}

func LoadNetConf(bytes []byte) (*NetConf, error) {
	n := &NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, fmt.Errorf("failed to load netconf: %s", err)
	}
	return n, nil
}
