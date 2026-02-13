#!/bin/bash
ssh admin@192.168.1.230 "\$env:HTTP_PROXY='http://192.168.6.117:7897'; \$env:HTTPS_PROXY='http://192.168.6.117:7897'; powershell -ExecutionPolicy Bypass -c "irm https://astral.sh/uv/install.ps1 | iex""
