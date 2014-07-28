#!/bin/sh
vzlist -a -H -o ctid | xargs -L1 -I {} sh -c 'vzctl stop {} --fast && vzctl delete {}'
brctl show | cut -f1 | tail -n +2 | xargs -L1 -I {} sh -c 'ifconfig {} down && brctl delbr {}'