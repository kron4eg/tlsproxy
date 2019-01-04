initca:
	cfssl gencert -initca ca-csr.json | cfssljson -bare ca

clean:
	rm -f *.pem
	rm -f *.csr
