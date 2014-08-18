#!/bin/bash
while true
do
	vzlist -S -H -o ctid | xargs -L1 -I {} sh -c 'vzctl start {}'
	sleep 10
done
