#!/usr/bin/env bash

DROPBOX_VAULT=~/Dropbox/1Password.opvault

VAULT=${OP_VAULT:=$DROPBOX_VAULT}
read -p "$VAULT Password: " -s PASSWORD
echo

./opvault-reader $VAULT $PASSWORD $1
