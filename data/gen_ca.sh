#!/usr/bin/env bash

CONFIG="
[req]
distinguished_name=dn
[dn]
[ext]
basicConstraints=CA:TRUE,pathlen:0
"

openssl req \
  -config <(echo "$CONFIG") \
  -subj "/CN=logi.im/O=LOGI, Inc." \
  -new \
  -newkey rsa:2048 \
  -nodes \
  -x509 \
  -days 365000 \
  -extensions ext \
  -keyout ca.key \
  -out ca.crt
