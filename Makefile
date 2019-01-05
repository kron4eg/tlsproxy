all: build

.PHONY: build
build:
	go build -v

.PHONY: run
run: build server-cert
	cd pki && \
	../tlsproxy

.PHONY: initca
initca: pki/ca.pem

.PHONY: server-cert
server-cert: pki/server.pem

.PHONY: client-cert
client-cert: pki/client.pem

pki/ca.pem:
	cfssl gencert -initca pki/ca-csr.json | cfssljson -bare pki/ca

pki/server.pem: pki/ca.pem
	cfssl gencert \
		-ca=pki/ca.pem \
		-ca-key=pki/ca-key.pem \
		-config=pki/ca-config.json \
		-profile=server \
		pki/server.json | \
			cfssljson -bare pki/server

pki/client.pem: pki/ca.pem
	cfssl gencert \
		-ca=pki/ca.pem \
		-ca-key=pki/ca-key.pem \
		-config=pki/ca-config.json \
		-profile=client \
		pki/client.json | \
			cfssljson -bare pki/client

clean:
	rm -f pki/*.pem
	rm -f pki/*.csr
