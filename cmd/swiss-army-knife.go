package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"swiss-army-knife-cni/pkg/sak"
	"swiss-army-knife-cni/pkg/types"
	"swiss-army-knife-cni/pkg/version"

	"github.com/containernetworking/cni/pkg/skel"
	cniTypes "github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/040"
	cniVersion "github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
)

func main() {
	skel.PluginMain(
		cmdAdd,
		nil,
		cmdDel,
		cniVersion.PluginSupports("0.1.0", "0.2.0", "0.3.0", "0.4.0"),
		"SwissArmyKnife CNI "+version.Version)
}

func cmdAdd(args *skel.CmdArgs) error {
	// We try to do as little as possible to get the annotation, and only do more if it has it.
	conf, err := types.LoadNetConf(args.StdinData)
	if err != nil {
		err = fmt.Errorf("Error parsing CNI configuration \"%s\": %s", args.StdinData, err)
		return err
	}

	cniresult, err := current.NewResultFromResult(conf.PrevResult)

	anno, err := sak.GetAnnotation(args, conf)
	if err != nil {
		return err
	}

	// We only do the rest if we have an annotation...
	if anno != "" {

		// Actually we wanna differentiate between docker & crio
		// crio: /var/run/netns/c00fc6c1-9a7e-4fb5-b415-be618736dec7
		// docker: /proc/27690/ns/net
		isCrio := strings.Contains(args.Netns, "/var/run/netns/")
		usenetns := args.Netns
		if !isCrio {
			usenetns, err = sak.BindDockerNetns(args.Netns, args.ContainerID)
			if err != nil {
				return err
			}
			defer sak.UnbindDockerNetns(usenetns)
			sak.WriteToSocket(fmt.Sprintf("!bang Mounted docker netns @ %s", usenetns), conf)
		}
		usenetns = strings.ReplaceAll(usenetns, "/var/run/netns/", "")

		// This worked with docker!
		// sudo ip netns 896d161cd20d60f239df27dbaa3f5d5f108ae7940390edc346d7445aa667ebcf ip addr

		// Convert the netns into a PID
		// I'm just stripping out non digits here.
		replaceiprx, _ := regexp.Compile(`\D`)
		pid := replaceiprx.ReplaceAllString(args.Netns, "")

		// Figure out the current interface name.
		// We get the last one in the list that has a sandbox
		// sak.WriteToSocket(fmt.Sprintf("!bang cniresult: %+v", cniresult.Interfaces), conf)
		currentInterface := ""
		for _, v := range cniresult.Interfaces {
			if v.Sandbox != "" {
				currentInterface = v.Name
			}
		}

		sak.WriteToSocket(fmt.Sprintf("!bang =========== isCrio: %v / pid: %s / ifname: %s / netns: %s", isCrio, pid, currentInterface, args.Netns), conf)
		// sak.WriteToSocket(fmt.Sprintf("!bang anno: %+v", anno), conf)
		commands, err := sak.ParseAnnotation(anno)
		if err != nil {
			sak.WriteToSocket(fmt.Sprintf("Error parsing command: %v", err), conf)
			return err
		}
		sak.WriteToSocket(fmt.Sprintf("Detected commands: %v", commands), conf)
		err = sak.ProcessCommands(usenetns, commands, conf)
		if err != nil {
			return err
		}
	}

	return cniTypes.PrintResult(cniresult, conf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) (err error) {
	netNS, err := ns.GetNS(args.Netns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting netNS: %s\n", err)
	}
	defer netNS.Close()
	return nil
}
