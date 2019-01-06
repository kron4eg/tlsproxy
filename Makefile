all: build

.PHONY: build
build:
	go build -v

.PHONY: run
run: build server-cert
	cd pki && \
	../tlsproxy

.PHONY: initca
initca:
	$(MAKE) -C ./pki initca

.PHONY: server-cert
server-cert:
	$(MAKE) -C ./pki server-cert

.PHONY: client-cert
client-cert:
	$(MAKE) -C ./pki client-cert

clean:
	$(MAKE) -C ./pki clean
