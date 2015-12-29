# Security settings of GoVideo

## Generate a self-signed certificate
```
// Generate server key pair
openssl genrsa -des3 -out server.key 2048

// Get separate private key from key pair.
openssl rsa -in server.key -out server.key

// Get certificate signing request
openssl req -sha256 -new -key server.key -out sign_request.csr -subj '/CN=135.55.22.67'

// Generate key and certificate files
openssl x509 -req -days 365 -in sign_request.csr -signkey server.key -out cert.crt

```

## Alternative one line way of gennerate the certificate
```
// This requires a pass phrase
// Use sha256 to avoid browser warning:
openssl req -x509 -sha256 -newkey rsa:2048 -keyout key.pem -out cert.pem -days XXX

// Use sha1 by default
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days XXX
```
XXX can be replaced with any days for this certificate to remain valid
