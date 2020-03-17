#!/bin/bash

read -p "Enter q3 server address including port (e.g. q3.example.com:27960): " q3_server_addr
echo
read -sp "Enter rcon password for the q3 server: " q3_password
echo
read -p "Enter telegram chat ID (in the format @channelusername): " telegram_chat_id
echo
read -sp "Enter telegram bot token: " telegram_token
echo

declare -A vars
vars=(
		[PWD]="$(pwd)/bin/q3serverbot"
		[Q3_SERVER_ADDR]=$q3_server_addr
		[Q3_PASSWORD]=$q3_password
		[TELEGRAM_CHAT_ID]=$telegram_chat_id
		[TELEGRAM_TOKEN]=$telegram_token
		[USER]=$USER
)

filename=$1

echo "Generating a unit file in $filename..."
cp "unit.tmpl" $filename

for key in "${!vars[@]}"
do
		search="%%$key%%"
		replace=${vars[$key]}
		sed -i "s|${search}|${replace}|g" "$filename"
done
echo "The unit file has been generated in $filename..."
