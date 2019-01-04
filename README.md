# tlsproxy
simple TLS proxy with obligatory client certificate validation

It's useful to organize authorized (via TLS client certificate validation) access to resources that natively don't support TLS client authentication (like prometheus node_exporter for example).
