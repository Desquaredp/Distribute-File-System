#!/bin/bash

# Check if the current user is dmistry4
if [ "$(whoami)" != "dmistry4" ]; then
    echo "This script can only be run by dmistry4"
    exit 1
fi

# Loop through hostnames orion01 to orion09
for host in orion{01..12}
do
    echo "Connecting to $host"
    # SSH into the host and execute the script
    ssh $host '
        # Kill all processes matching the pattern /tmp/go-build*/b001/exe/spawn
        pkill -f "/tmp/go-build.*\/b001\/exe\/spawn"

        # Check if the process was killed
        if ps -p $(pgrep -f "/tmp/go-build.*\/b001\/exe\/spawn") > /dev/null; then
            echo "Failed to kill the spawn process on $host"
        else
            echo "Successfully killed the spawn process on $host"
        fi
    '
done