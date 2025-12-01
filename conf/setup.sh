#!/bin/bash

set -euo pipefail

echo "Linking sysusers config..."

mkdir -p /etc/sysusers.d

if [ ! -f /etc/sysusers.d/widgets.conf ]; then
    ln -s "/var/widgets.shrt.day/conf/widgets.conf" /etc/sysusers.d/widgets.conf
fi

echo "Creating user..."
systemd-sysusers

echo "Linking unit..."
rm /etc/systemd/system/widgets.service

systemctl link "/var/widgets.shrt.day/conf/widgets.service"

echo "Reloading daemon..."
systemctl daemon-reload
systemctl enable widgets

echo "Fixing initial permissions..."
chown -R widgets:widgets "/var/widgets.shrt.day"

find "/var/widgets.shrt.day" -type d -exec chmod 755 {} +
find "/var/widgets.shrt.day" -type f -exec chmod 644 {} +

chmod +x "/var/widgets.shrt.day/widgets"

echo "Setup complete, starting service..."

service widgets start

echo "Done."
