#!/bin/bash

# Script deploy nhanh lên VPS
# Sử dụng: ./deploy.sh user@vps-ip:/opt/qh88

if [ -z "$1" ]; then
    echo "Sử dụng: ./deploy.sh user@vps-ip:/path/to/deploy"
    echo "Ví dụ: ./deploy.sh root@192.168.1.100:/opt/qh88"
    exit 1
fi

DEPLOY_PATH=$1

echo "🚀 Bắt đầu deploy QH88..."

# Build binary
echo "📦 Building binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o qh88-server main.go

if [ $? -ne 0 ]; then
    echo "❌ Build thất bại!"
    exit 1
fi

echo "✅ Build thành công!"

# Upload files
echo "📤 Uploading files..."
scp qh88-server "$DEPLOY_PATH/"
scp -r web/ "$DEPLOY_PATH/"
scp docker-compose.yml "$DEPLOY_PATH/" 2>/dev/null || true
scp Dockerfile "$DEPLOY_PATH/" 2>/dev/null || true

echo "✅ Upload thành công!"

# SSH và restart service
echo "🔄 Restarting service..."
ssh "${DEPLOY_PATH%%:*}" << EOF
cd ${DEPLOY_PATH#*:}
chmod +x qh88-server
if command -v docker-compose &> /dev/null; then
    docker-compose down
    docker-compose up -d --build
else
    sudo systemctl restart qh88 || echo "Chạy: ./qh88-server"
fi
EOF

echo "✅ Deploy hoàn tất!"
echo "🌐 Truy cập: http://${DEPLOY_PATH%%:*} | grep -oP '\\d+\\.\\d+\\.\\d+\\.\\d+' || echo 'your-vps-ip'}:8080"

