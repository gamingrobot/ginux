#!/bin/sh
CID="$1"
NETWORKID="$2"

vzctl set $CID --netif_add eth$NETWORKID --save
echo 0 >"/proc/sys/net/ipv4/conf/veth$CID.$NETWORKID/forwarding"
echo 0 >"/proc/sys/net/ipv4/conf/veth$CID.$NETWORKID/proxy_arp"
vzctl exec $CID ifconfig eth$NETWORKID 0
vzctl exec $CID ip addr add 192.168.$NETWORKID.$CID/24 dev eth$NETWORKID