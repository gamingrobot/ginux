#!/bin/bash
echo "I am the deploy script that is ran on ginux-dev"
killall game_server
cd /home/ginux/ginux/
cd server/game_server
screen -dmSL GS ./game_server