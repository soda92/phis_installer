#!/bin/bash
HOST="admin@192.168.1.230"
REMOTE_DIR="C:/installer_test"
PS1_SCRIPT="server_actions/benchmark.ps1"

# Find latest installer (by modification time)
# Matches pattern: build/数字员工平台v*.exe
ORIGINAL_INSTALLER=$(ls -t build/数字员工平台v*.exe 2>/dev/null | head -n1)

if [ -z "$ORIGINAL_INSTALLER" ]; then
    echo "Error: No installer found matching 'build/数字员工平台v*.exe'"
    exit 1
fi

echo "Using latest installer: $ORIGINAL_INSTALLER"

TEMP_INSTALLER="build/installer.exe"

echo "Renaming installer to ASCII for transfer..."
cp "$ORIGINAL_INSTALLER" "$TEMP_INSTALLER"

echo "Transferring files to $HOST..."
ssh $HOST "mkdir $REMOTE_DIR 2> nul"
scp "$TEMP_INSTALLER" "$PS1_SCRIPT" "$HOST:$REMOTE_DIR/"

echo "Running benchmark on remote server..."
ssh $HOST "powershell -ExecutionPolicy Bypass -File $REMOTE_DIR/benchmark.ps1"

# Cleanup
rm "$TEMP_INSTALLER"
