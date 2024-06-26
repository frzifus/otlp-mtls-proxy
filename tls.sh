#!/bin/bash

mkdir -p tls

# generate root cert
openssl genpkey -algorithm RSA -out tls/rootCA.key
openssl req -x509 -new -nodes -key tls/rootCA.key -sha256 -days 365 -out tls/rootCA.crt -subj "/CN=My Root CA"

# generate server key
openssl genpkey -algorithm RSA -out tls/server.key
# Create a Certificate Signing Request (CSR) for the server
openssl req -new -key tls/server.key -out tls/server.csr -subj "/CN=localhost.localdomain"
# Sign the server CRS with CA
openssl x509 -req -in tls/server.csr -CA tls/rootCA.crt -CAkey tls/rootCA.key -CAcreateserial -out tls/server.crt -days 365 -sha256 -extfile <(printf "subjectAltName=DNS:localhost,DNS:localhost.localdomain")


# generate client key
openssl genpkey -algorithm RSA -out tls/client.key
# Create a Certificate Signing Request (CSR) for the client
openssl req -new -key tls/client.key -out tls/client.csr -subj "/CN=Client"
# Sign the client CSR with your CA
openssl x509 -req -in tls/client.csr -CA tls/rootCA.crt -CAkey tls/rootCA.key -CAcreateserial -out tls/client.crt -days 365 -sha256
