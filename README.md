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

## Installation

This is the recommended way to install the tool. You can also download the binary from the releases page.

```bash
wget https://github.com/lukaszmoskwa/docknat/releases/latest/download/docknat && chmod +x docknat && sudo mv docknat /usr/bin
```

### Run the job as a systemd service

Create a systemd service file in `/etc/systemd/system/docknat.service`:

```bash
sudo vim /etc/systemd/system/docknat.service
```

Add the following content (update the `User` field with your username):

```bash
[Unit]
Description=Docknat
After=network.target docker.service

[Service]
ExecStart=/usr/bin/docknat start
Restart=always
User=<your-user>
Group=docker
# Grant permissions to access Docker and iptables
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_RAW CAP_SYS_MODULE
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_RAW CAP_SYS_MODULE
NoNewPrivileges=true
# Ensure the service has access to the Docker socket
BindPaths=/var/run/docker.sock
# Ensure the service has access to iptables
# This can be more complex; adjust as necessary for your specific case
# Also consider that iptables requires elevated permissions
ProtectSystem=full
ProtectHome=yes
PrivateDevices=yes

[Install]
WantedBy=multi-user.target
```

Then reload the systemd daemon and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now docknat
```

And start the service:

```bash
sudo systemctl start docknat
```

You can check the status of the service with:

```bash
sudo systemctl status docknat
```

And read the logs with:

```bash
journalctl -u docknat
```

### IPTables rules

The tool will write the following rules in the `nat` table:

```bash
-A PREROUTING -i <interface> -j DNAT --to-destination <container-ip>
```

And you can check them with

```bash
sudo iptables -t nat -L
```

## Dev Setup

```bash
git clone github.com/lukaszmoskwa/docknat
cd docknat
go mod download
```

## TODO

- [ ] Improve documentation
- [ ] Add some tests
