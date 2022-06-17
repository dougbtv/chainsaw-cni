package sak

import (
  "context"
  "encoding/json"
  "fmt"
  cnitypes "github.com/containernetworking/cni/pkg/types"
  "net"
  "os"
  "os/exec"
  "regexp"
  "strings"
  // current "github.com/containernetworking/cni/pkg/types/040"
  // cniVersion "github.com/containernetworking/cni/pkg/version"
  "github.com/containernetworking/cni/pkg/skel"
  "swiss-army-knife-cni/pkg/types"
  "time"

  "k8s.io/client-go/kubernetes"
  // "k8s.io/client-go/kubernetes/scheme"
  // v1core "k8s.io/client-go/kubernetes/typed/core/v1"
  "k8s.io/client-go/rest"
  "k8s.io/client-go/tools/clientcmd"
  // "k8s.io/client-go/tools/record"
  v1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
  sakAnnotation = "k8s.v1.cni.cncf.io/swiss-army-knife"
)

func runCommand(command string, argsstring string) (string, error) {
  args := strings.Fields(argsstring)
  cmd := exec.Command(command, args...)
  // cmd.Stdin = strings.NewReader("some input")
  out, err := cmd.CombinedOutput()
  if err != nil {
    return string(out), err
  }
  return string(out), nil
}

// BindDockerNetns bind mounts the docker netns so `ip netns` can use it
func BindDockerNetns(originalNetnsPath string, containerid string) (string, error) {
  newNetnsPath := "/var/run/netns/" + containerid
  _, err := os.Stat(newNetnsPath)
  if os.IsNotExist(err) {
    file, err := os.Create(newNetnsPath)
    if err != nil {
      return "", fmt.Errorf("Failed to create file: %s", newNetnsPath)
    }
    file.Close()
  } else {
    return "", fmt.Errorf("The target netns file exists: %s", newNetnsPath)
  }

  // Now let's bind mount the original to this.
  // mount -o bind /proc/3357/ns/net /var/run/netns/$container_id
  _, err = runCommand("mount", fmt.Sprintf("-o bind %s %s", originalNetnsPath, newNetnsPath))
  if err != nil {
    return "", fmt.Errorf("Failed to mount %s to %s: %s", originalNetnsPath, newNetnsPath, err)
  }
  // fmt.Printf("output: %q\n", out.String())

  return newNetnsPath, nil
}

// UnbindDockerNetns unmounts our bind for a docker ns.
func UnbindDockerNetns(dockernetns string) error {
  _, err := runCommand("umount", dockernetns)
  if err != nil {
    return fmt.Errorf("Failed to umount %s: %s", dockernetns, err)
  }

  err = os.Remove(dockernetns)
  if err != nil {
    return fmt.Errorf("Failed to remove %s: %s", dockernetns, err)
  }
  return nil
}

// ProcessCommands runs all of the intended commands from the pod annotation
func ProcessCommands(netns string, commands []string, conf *types.NetConf) error {
  for _, v := range commands {
    // sudo ip netns 896d161cd20d60f239df27dbaa3f5d5f108ae7940390edc346d7445aa667ebcf ip addr
    argstring := fmt.Sprintf("netns exec %s ip %s", netns, v)
    debugcmdstring := "ip " + argstring
    output, err := runCommand("ip", argstring)
    WriteToSocket(fmt.Sprintf("Running %s ===============\n%s\n", debugcmdstring, output), conf)
    if err != nil {
      WriteToSocket(fmt.Sprintf("Command '%s' failed with: %s", debugcmdstring, err), conf)
      return fmt.Errorf("Command '%s' failed with: %s", debugcmdstring, err)
    }
  }
  return nil
}

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

// ParseAnnotation parses JSON out of the annotation
func ParseAnnotation(rawannotation string) ([]string, error) {

  var commands []string

  // Parse it if we have JSON.
  if strings.Contains(rawannotation, "[") {
    if err := json.Unmarshal([]byte(rawannotation), &commands); err != nil {
      return nil, fmt.Errorf("failed to parse JSON annotation: %s", err)
    }
  } else {
    // Just parse it as a command.
    commands = append(commands, rawannotation)
  }

  // Cycle through each command and make sure it's legit.
  //
  validationrx, _ := regexp.Compile("^[^\\.\\/][\\w\\s\\.\\:_\\-\\d\\/]+$")
  replaceiprx, _ := regexp.Compile("^\\s*?ip\\s+")
  // r.MatchString("peach")
  for idx, v := range commands {
    if !validationrx.MatchString(v) {
      return nil, fmt.Errorf("We cannot validate the value: '%s' (it's validated like this: https://regex101.com/r/vPKuZC/1)", v)
    }

    // You can use the "ip" name optionally, but we don't want to use the user input.
    if replaceiprx.MatchString(v) {
      commands[idx] = replaceiprx.ReplaceAllString(v, "")
    }
  }

  return commands, nil

}

// GetAnnotation gets a pod annotation
func GetAnnotation(args *skel.CmdArgs, conf *types.NetConf) (string, error) {
  kubeClient, err := GetK8sClient(conf.Kubeconfig, nil)
  if err != nil {
    return "", fmt.Errorf("error getting k8s client: %v", err)
  }

  k8sArgs, err := GetK8sArgs(args)
  if err != nil {
    return "", fmt.Errorf("error getting k8s args: %v", err)
  }

  err = WriteToSocket(fmt.Sprintf("!bang k8sArgs: %+v", k8sArgs), conf)
  if err != nil {
    return "", err
  }

  pod, err := getPod(kubeClient, k8sArgs)
  if err != nil {
    return "", err
  }

  sakannovalue := pod.Annotations[sakAnnotation]
  return sakannovalue, nil

}

func getPod(kubeClient *ClientInfo, k8sArgs *types.K8sArgs) (*v1.Pod, error) {
  if kubeClient == nil {
    return nil, nil
  }

  podNamespace := string(k8sArgs.K8S_POD_NAMESPACE)
  podName := string(k8sArgs.K8S_POD_NAME)
  // podUID := string(k8sArgs.K8S_POD_UID)

  pod, err := kubeClient.GetPod(podNamespace, podName)
  if err != nil {
    return nil, fmt.Errorf("error getting pod: %v", err)
  }

  return pod, nil
}

// GetK8sArgs gets k8s related args from CNI args
func GetK8sArgs(args *skel.CmdArgs) (*types.K8sArgs, error) {
  k8sArgs := &types.K8sArgs{}

  err := cnitypes.LoadArgs(args.Args, k8sArgs)
  if err != nil {
    return nil, err
  }

  return k8sArgs, nil
}

// ClientInfo contains information given from k8s client
type ClientInfo struct {
  Client kubernetes.Interface
  // NetClient        netclient.K8sCniCncfIoV1Interface
  // EventBroadcaster record.EventBroadcaster
  // EventRecorder    record.EventRecorder
}

// GetPod gets pod from kubernetes
func (c *ClientInfo) GetPod(namespace, name string) (*v1.Pod, error) {
  return c.Client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetK8sClient gets client info from kubeconfig
func GetK8sClient(kubeconfig string, kubeClient *ClientInfo) (*ClientInfo, error) {
  // logging.Debugf("GetK8sClient: %s, %v", kubeconfig, kubeClient)
  // If we get a valid kubeClient (eg from testcases) just return that
  // one.
  if kubeClient != nil {
    return kubeClient, nil
  }

  var err error
  var config *rest.Config

  // Otherwise try to create a kubeClient from a given kubeConfig
  if kubeconfig != "" {
    // uses the current context in kubeconfig
    config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
      return nil, fmt.Errorf("GetK8sClient: failed to get context for the kubeconfig %v: %v", kubeconfig, err)
    }
  } else if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
    // Try in-cluster config where multus might be running in a kubernetes pod
    config, err = rest.InClusterConfig()
    if err != nil {
      return nil, fmt.Errorf("GetK8sClient: failed to get context for in-cluster kube config: %v", err)
    }
  } else {
    // No kubernetes config; assume we shouldn't talk to Kube at all
    return nil, nil
  }

  // Specify that we use gRPC
  config.AcceptContentTypes = "application/vnd.kubernetes.protobuf,application/json"
  config.ContentType = "application/vnd.kubernetes.protobuf"
  // Set the config timeout to one minute.
  config.Timeout = time.Minute

  // creates the clientset
  client, err := kubernetes.NewForConfig(config)
  if err != nil {
    return nil, err
  }

  return &ClientInfo{
    Client: client,
  }, nil
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
