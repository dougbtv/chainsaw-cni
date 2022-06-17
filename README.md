# chainsaw-cni

Chainsaw: A configuration and debugging tool for rough cuts using CNI chains.

The gist is that it allows you to tweak parameters of your network namespaces at runtime. Enables you to run `ip` commands against your containers network namespace from within a CNI chain.

You add chainsaw-cni as a member of a CNI chain, then... you annotate a pod -- you use `ip` commands, and it modifies your network namespace using the `ip` command.

## Installation

## Example Usage

## Disclaimers

This might start out with some considerations that you want to take seriously as an administrator. There's a non-zero probability that someone can do something nasty to your network if they can annotate pods and you're using this tool.

It's a chainsaw after all, [use it carefully](http://www.gameoflogging.com/).

## TODO

Filter expressions: Limit to just a subset of `ip` commands.