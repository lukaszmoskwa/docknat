# Docknat

Docknat is a utility tool that monitors running docker containers and updates the NAT rules in iptables in order to allow smooth communication with other containers on custom interfaces.

## Why

When you try to make different system work with each other, things can get pretty complex pretty fast. And the worst issues are always somehow related to networking.
This is the case when you want to use Consul and Nomad while using a Tailscale interface and Docker containers.

Each of these softwares manage some network rules and make it hard to make them work together.
In particular, Docker and TailScale are not always friends and you need to make sure that the NAT rules are correctly in order for packets to find their way to the right container.
Docker writes its own rules in the `DOCKER` chain of the `nat` table, while Tailscale writes its own rules in the `TS-INPUT` chain of the `filter` table.

At this point, it's only fair game to write a utility that writes some other rules in the `nat` table.

### More about the issue

To reproduce the issue, you can follow these steps:

- Start a clean Ubuntu 22.04 VM on any cloud provider
- Install Docker and Tailscale (make sure you restart the docker daemon after installing Tailscale, the DNS resolution will not work otherwise since by default Docker uses the host's DNS)
- Create a consul server. Ensure that you are advertising the correct IP address (the Tailscale IP address) and not the public IP address. You don't want your applications to be visible from the outside world.
- Create a nomad server and client (for the sake of the example, you can have server and client on the same node)
- On another machine connected to the same Tailscale network, create a nomad client, consul client and connect it to the consul cluster
- Run containers on both nomad clients

At this point, you will notice that the containers on the different clients can communicate, but they fail to communicate with containers on the same node. This madness is due to some conflicting rules in the iptables.

### Extra read

- Docker ignores UFW (https://askubuntu.com/questions/652556/uncomplicated-firewall-ufw-is-not-blocking-anything-when-using-docker)

## Scope

The scope of the project is limited to solving this issue and nothing more. I'm not planning to extend it unless I find other issues that need to be solved and cases that can benefit from this tool.

## Dev Setup

```bash
git clone github.com/lukaszmoskwa/docknat
cd docknat
go mod download
```
