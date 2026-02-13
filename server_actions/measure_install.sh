#!/bin/bash
HOST="admin@192.168.1.230"
REMOTE_DIR="C:/installer_test"
ORIGINAL_INSTALLER="数字员工平台20.exe"
LOCAL_PATH="build/$ORIGINAL_INSTALLER"
TEMP_INSTALLER="build/installer.exe"
PS1_SCRIPT="server_actions/benchmark.ps1"

echo "Renaming installer to ASCII for transfer..."
cp "$LOCAL_PATH" "$TEMP_INSTALLER"

echo "Transferring files to $HOST..."
ssh $HOST "mkdir $REMOTE_DIR 2> nul"
scp "$TEMP_INSTALLER" "$PS1_SCRIPT" "$HOST:$REMOTE_DIR/"

echo "Running benchmark on remote server..."
ssh $HOST "powershell -ExecutionPolicy Bypass -File $REMOTE_DIR/benchmark.ps1"

# Cleanup
rm "$TEMP_INSTALLER"
