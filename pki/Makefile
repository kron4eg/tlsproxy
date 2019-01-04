.PHONY: initca
initca: ca.pem
server-cert: server.pem
client-cert: client.pem

ca.pem:
	cfssl gencert -initca ca-csr.json | cfssljson -bare ca

server.pem: ca.pem
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=ca-config.json \
		-profile=server \
		server.json | \
			cfssljson -bare server

client.pem: ca.pem
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=ca-config.json \
		-profile=client \
		client.json | \
			cfssljson -bare client

clean:
	rm -f *.pem
	rm -f *.csr
