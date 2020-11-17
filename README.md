# bgp-lb

**disclaimer**: The project is on very early and experimental stages!

Simple app that advertises a service ip to a list of bgp peers based on the
status of a healthcheck. The aim is to advertise the same service ip via
multiple hosts and succeed load balancing using bgp multipath on the bgp peers
(network routers)

Table of Contents
=================

   * [bgp-lb](#bgp-lb)
      * [Functionality](#functionality)
      * [Considerations](#considerations)
      * [Configuration](#configuration)
         * [BGP](#bgp)
         * [Service - Healthchecks](#service---healthchecks)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc)

## Functionality

The apps performs the following tasks:

- Makes sure a a dummy interface named after the service exists in the node.
- Binds the service ip address to the dummy interface.
- Creates an IPVS virtual service for the service ip and adds the local service
  target as destination (uses the router address and the local target port
  provided as configuration to create the destination)
- Starts a bgp server and configures a list of given peers.
- Periodically checks the defined healthcheck and adds or removes a path to the
  service via the host on the bgp server respectively.

As a result, when the check is healthy the node advertises the service ip with
it's own address as the next hop.

## Considerations

- The app needs to establish BGP peering session with your network routers.
- Routers should be configured to support multipath
- Load balancing would happen via the routers routing table based on the path
  selection algorithm (usually round robin)

## Configuration

An example of the full supported configuration can be found [here](./config-example.json)

### BGP

A list of peers (ip address/as number) can be specified so that the bgp server
will try to establish bgp connections. For example:
```
    "peers": [
      {
        "address": "10.88.0.253",
	"as": 65512
      },
      {
        "address": "10.88.0.254",
	"as": 65512
      }
    ]
```

For the local server the app expects configuration for the router id (an ip that
can route traffic to the host on the network), the local as number and a listen
port.
```
    "local": {
      "routerID": "10.88.0.200",
      "as": 65512,
      "listenPort": -1
    }

```

### Service - Healthchecks

Currently the app expects a very simple http health check that checks for 2XX
response code from a service running on a local port.
Example:
```
    "httphealthcheck": {
       "port": 8080
    }
```
