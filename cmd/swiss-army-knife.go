package main

import (
	"fmt"
	"os"
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
	conf, err := types.LoadNetConf(args.StdinData)
	if err != nil {
		err = fmt.Errorf("Error parsing CNI configuration \"%s\": %s", args.StdinData, err)
		return err
	}

	err = sak.WriteToSocket(fmt.Sprintf("!bang netns: %s", conf.Foo, args.Netns), conf)
	if err != nil {
		return err
	}

	anno, err := sak.GetAnnotation(args, conf)
	if err != nil {
		return err
	}

	// We only do the rest if we have an annotation...
	if anno != "" {
		// sak.WriteToSocket(fmt.Sprintf("!bang anno: %+v", anno), conf)
		commands, err := sak.ParseAnnotation(anno)
		if err != nil {
			sak.WriteToSocket(fmt.Sprintf("Error parsing command: %v", err), conf)
			return err
		}
		sak.WriteToSocket(fmt.Sprintf("!bang Detected commands: %v", commands), conf)
	}

	result, err := current.NewResultFromResult(conf.PrevResult)

	return cniTypes.PrintResult(result, conf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) (err error) {
	netNS, err := ns.GetNS(args.Netns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting netNS: %s\n", err)
	}
	defer netNS.Close()
	return nil
}
