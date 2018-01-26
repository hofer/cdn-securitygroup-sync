#!/bin/bash
set -e

wget -c -O $1/main.zip https://github.com/hofer/cdn-securitygroup-sync/releases/download/v1.0.2/v1.0.2.zip

jq -n '{"filename":"main.zip"}'