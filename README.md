# tlsproxy
Simple TLS proxy with obligatory client certificate validation

It's useful to organize authorized (via TLS client certificate validation)
access to resources that natively don't support TLS client authentication (like
prometheus node_exporter for example).

## Config

Following will dump default config to stdout

```shell
tlsproxy -genconfig
```

Run it with custom config
```shell
tlsproxy -config config.json
```

### Default Config

```json
{
  "listen": ":19100",               # port to listen on
  "required_client_name": "",       # in case if client name is needed to be verified
  "tls": {                          # default TLS server config
    "ca": "ca.pem",
    "cert": "server.pem",
    "key": "server-key.pem"
  },
  "vhosts": {
    "": 9100                        # proxy upstream target port. host is always 127.0.0.1
  }
}
```

## PKI
Example PKI configuration for `cfssl` tool is located in `/pki` directory. To
quickly generate CA/server/client certificates please use 
`make server-cert client-cert`.
