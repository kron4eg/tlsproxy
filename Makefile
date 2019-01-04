all: build

build:
	go build -v

.PHONY: initca
initca: ca.pem

.PHONY: server-cert
server-cert: server.pem

.PHONY: client-cert
client-cert: client.pem

ca.pem:
	cfssl gencert -initca pki/ca-csr.json | cfssljson -bare ca

server.pem: ca.pem
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=pki/ca-config.json \
		-profile=server \
		pki/server.json | \
			cfssljson -bare server

client.pem: ca.pem
	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=pki/ca-config.json \
		-profile=client \
		pki/client.json | \
			cfssljson -bare client

clean:
	rm -f *.pem
	rm -f *.csr
