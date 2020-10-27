#!/usr/bin/env bash

OBJ=extra/logi.crt
SAVE_PATH=/usr/share/ca-certificates/$OBJ
CONFIG_PATH=/etc/ca-certificates.conf

cp ca.crt "$SAVE_PATH"
if ! grep -q $OBJ $CONFIG_PATH; then
  echo $OBJ >> $CONFIG_PATH
fi

update-ca-certificates --fresh
