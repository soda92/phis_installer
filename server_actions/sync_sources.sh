#!/bin/bash
tar --exclude='__pycache__' -cf phis_sources.tar installer resources manage.py
scp phis_sources.tar admin@192.168.1.230:C:/installer/
ssh admin@192.168.1.230 "cd C:\installer && tar -xf phis_sources.tar"
rm phis_sources.tar
