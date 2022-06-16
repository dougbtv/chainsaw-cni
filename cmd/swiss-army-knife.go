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

	fmt.Fprintf(os.Stderr, "!bang value of foo: %s\n | netns: %s", conf.Foo, args.Netns)
	err = sak.WriteToSocket(fmt.Sprintf("!bang value of foo: %s | netns: %s", conf.Foo, args.Netns), conf)
	if err != nil {
		return err
	}

	sak.GetAnnotation(args, conf)

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
