#!/usr/bin/env bash


key_dir="./tmp"
mkdir "$key_dir"
chmod 0700 "$key_dir"
cd "$key_dir"

cat >server.conf <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
prompt = no
[req_distinguished_name]
CN = sigrun.default.svc
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = sigrun.default.svc
EOF

# Generate the CA cert and private key
openssl req -nodes -new -x509 -keyout ca.key -out ca.crt -subj "/CN=Admission Controller Webhook Demo CA"
# Generate the private key for the webhook server
openssl genrsa -out wh.key 2048
# Generate a Certificate Signing Request (CSR) for the private key, and sign it with the private key of the CA.
openssl req -new -key wh.key -subj "/CN=sigrun.default.svc" -config server.conf \
    | openssl x509 -req -CA ca.crt -CAkey ca.key -CAcreateserial -out wh.crt -extensions v3_req -extfile server.conf

ca_pem_b64="$(openssl base64 -A <"./ca.crt")"
printf "\n\nca cert\n\n"
echo "$ca_pem_b64"
printf "\n\nwh cert\n\n"
cat ./wh.crt | base64 | tr -d '\n'
printf "\n\nwh key\n\n"
cat ./wh.key | base64 | tr -d '\n'
printf "\n\n"

cd .. && rm -R tmp