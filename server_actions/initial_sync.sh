#!/bin/bash
tar --exclude='build' --exclude='.venv' --exclude='.git' --exclude='__pycache__' -cf phis_installer.tar .
scp phis_installer.tar admin@192.168.1.230:C:/installer/
ssh admin@192.168.1.230 "cd C:\installer && tar -xf phis_installer.tar"
rm phis_installer.tar
