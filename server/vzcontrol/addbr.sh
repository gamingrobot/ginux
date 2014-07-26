#!/bin/sh
NETWORKID=$1

brctl addbr vzbr$NETWORKID
ifconfig vzbr$NETWORKID 0
echo 0 >"/proc/sys/net/ipv4/conf/vzbr$NETWORKID/forwarding"
echo 0 >"/proc/sys/net/ipv4/conf/vzbr$NETWORKID/proxy_arp"