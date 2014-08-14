#!/bin/bash
echo "I am the deploy script that is ran on ginux-dev as root"
killall vzcontrol
cd /home/ginux/
cd server/vzcontrol
screen -dmSL VZC ./vzcontrol