package sak

import (
  "fmt"
  "net"
  "os"
  // cniTypes "github.com/containernetworking/cni/pkg/types"
  // current "github.com/containernetworking/cni/pkg/types/040"
  // cniVersion "github.com/containernetworking/cni/pkg/version"
  "swiss-army-knife-cni/pkg/types"
)

// WriteToSocket writes to our socketfile, for logging.
func WriteToSocket(output string, conf *types.NetConf) error {
  if conf.SocketEnabled {

    filestat, err := os.Stat(conf.SocketPath)
    if err != nil {
      return fmt.Errorf("socket file stat failed: %v", err)
    }

    if !filestat.IsDir() {
      if filestat.Mode()&os.ModeSocket == 0 {
        return fmt.Errorf("is not a socket file: %v", err)
      }
    }

    fmt.Fprintf(os.Stderr, "!bang output: %s\n", output)
    handler, err := net.Dial("unix", conf.SocketPath)
    if err != nil {
      return fmt.Errorf("can't open unix socket %v: %v", conf.SocketPath, err)
    }
    defer handler.Close()

    _, err = handler.Write([]byte(output + "\n"))
    if err != nil {
      return fmt.Errorf("socket write error: %v", err)
    }
  }
  return nil
}

// // LoadNetConf parses our cni configuration
// func LoadNetConf(bytes []byte) (*NetConf, error) {
//   conf := NetConf{}
//   if err := json.Unmarshal(bytes, &conf); err != nil {
//     return nil, fmt.Errorf("failed to load netconf: %s", err)
//   }

//   // Parse previous result
//   if conf.RawPrevResult != nil {
//     resultBytes, err := json.Marshal(conf.RawPrevResult)
//     if err != nil {
//       return nil, fmt.Errorf("could not serialize prevResult: %v", err)
//     }

//     res, err := cniVersion.NewResult(conf.CNIVersion, resultBytes)

//     if err != nil {
//       return nil, fmt.Errorf("could not parse prevResult: %v", err)
//     }

//     conf.RawPrevResult = nil
//     conf.PrevResult, err = current.NewResultFromResult(res)
//     if err != nil {
//       return nil, fmt.Errorf("could not convert result to current version: %v", err)
//     }
//   }

//   return &conf, nil
// }
