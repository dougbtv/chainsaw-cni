# swiss-army-knife-cni

It's your multi-tool for manipulating network namespaces in CNI chains. It slices, it dices... it juiliennes.

The gist is that it allows you to tweak parameters of your network namespaces at runtime. 

You add swiss-army-knife (aka SAK) as a member of a CNI chain, then... you annotate a pod -- you use `ip` commands, and it modifies your network namespace using the `ip` command.

## Installation

## Usage

## Disclaimers

This might start out with some considerations that you want to take seriously as an administrator. There's a non-zero probability that someone can do something nasty to your network if they can annotate pods and you're using this tool.

It's a knife, it's sharp, use it carefully.

## TODO

Filter expressions: Limit to just a subset of `ip` commands.