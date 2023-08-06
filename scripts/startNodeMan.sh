#!/bin/bash

for i in {01..12}; do
  ssh orion$i 'cd P2-extra-pickle-rick/src; go run node_manager/node_manager.go node_manager/resourceman_conn.go node_manager/nodeman_conn.go' &
done

echo "Spawned processes in the background."