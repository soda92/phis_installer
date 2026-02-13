#!/bin/bash
ssh admin@192.168.1.230 "cd C:\installer && \$env:HTTP_PROXY='http://192.168.6.117:7897'; \$env:HTTPS_PROXY='http://192.168.6.117:7897'; uv run manage.py build-upgrade --from-ver 1.9"
