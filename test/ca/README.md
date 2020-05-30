Generate certs for testing

Uses [cfssl](https://github.com/cloudflare/cfssl)

### Generate CA
```bash
cfssl genkey -initca ca-template.json | cfssljson -bare ca
```

### Generate Server Certs
```bash
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=server server-csr.json | cfssljson -bare server
```
