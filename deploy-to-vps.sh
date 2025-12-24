#!/bin/bash

# Script deploy QH88 lên VPS 72.62.120.215
# Sử dụng: bash deploy-to-vps.sh

VPS_IP="72.62.120.215"
VPS_USER="root"
VPS_PATH="/opt/qh88"
DOMAIN="qh88h1.com"

echo "🚀 Bắt đầu deploy QH88 lên VPS $VPS_IP..."

# Kiểm tra kết nối SSH
echo "📡 Kiểm tra kết nối SSH..."
if ! ssh -o ConnectTimeout=5 -o BatchMode=yes "$VPS_USER@$VPS_IP" exit 2>/dev/null; then
    echo "❌ Không thể kết nối SSH đến $VPS_USER@$VPS_IP"
    echo "💡 Đảm bảo bạn đã cấu hình SSH key hoặc có thể nhập password"
    exit 1
fi
echo "✅ Kết nối SSH thành công!"

# Build binary (nếu cần)
if [ ! -f "qh88-server" ]; then
    echo "📦 Building binary cho Linux..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o qh88-server main.go
    if [ $? -ne 0 ]; then
        echo "❌ Build thất bại!"
        exit 1
    fi
    echo "✅ Build thành công!"
fi

# Upload files
echo "📤 Uploading files lên VPS..."
ssh "$VPS_USER@$VPS_IP" "mkdir -p $VPS_PATH"

# Upload các file cần thiết
scp qh88-server "$VPS_USER@$VPS_IP:$VPS_PATH/" 2>/dev/null || echo "⚠️  qh88-server không tìm thấy, sẽ build trên VPS"
scp -r web/ "$VPS_USER@$VPS_IP:$VPS_PATH/"
scp docker-compose.yml "$VPS_USER@$VPS_IP:$VPS_PATH/" 2>/dev/null || true
scp Dockerfile "$VPS_USER@$VPS_IP:$VPS_PATH/" 2>/dev/null || true
scp nginx.conf "$VPS_USER@$VPS_IP:$VPS_PATH/" 2>/dev/null || true
scp setup-domain.sh "$VPS_USER@$VPS_IP:$VPS_PATH/" 2>/dev/null || true
scp go.mod go.sum "$VPS_USER@$VPS_IP:$VPS_PATH/" 2>/dev/null || true
scp main.go "$VPS_USER@$VPS_IP:$VPS_PATH/" 2>/dev/null || true

echo "✅ Upload thành công!"

# Deploy trên VPS
echo "🔧 Đang cấu hình trên VPS..."
ssh "$VPS_USER@$VPS_IP" << 'ENDSSH'
cd /opt/qh88

# Tạo thư mục data nếu chưa có
mkdir -p data

# Tạo file .env nếu chưa có
if [ ! -f .env ]; then
    echo "DOWNLOAD_TOKEN=tok123" > .env
    echo "✅ Đã tạo file .env"
fi

# Nếu có Docker, dùng Docker
if command -v docker-compose &> /dev/null || command -v docker compose &> /dev/null; then
    echo "🐳 Sử dụng Docker..."
    chmod +x setup-domain.sh 2>/dev/null || true
    
    # Build và chạy
    if command -v docker-compose &> /dev/null; then
        docker-compose down 2>/dev/null || true
        docker-compose up -d --build
    else
        docker compose down 2>/dev/null || true
        docker compose up -d --build
    fi
    echo "✅ Docker container đã được khởi động!"
else
    echo "📦 Sử dụng binary trực tiếp..."
    
    # Build trên VPS nếu chưa có binary
    if [ ! -f "qh88-server" ]; then
        if command -v go &> /dev/null; then
            echo "🔨 Building binary trên VPS..."
            CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o qh88-server main.go
        else
            echo "❌ Go chưa được cài đặt trên VPS!"
            echo "💡 Cài đặt Go hoặc upload file qh88-server"
            exit 1
        fi
    fi
    
    chmod +x qh88-server
    
    # Tạo systemd service
    cat > /tmp/qh88.service << 'EOF'
[Unit]
Description=QH88 Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/qh88
ExecStart=/opt/qh88/qh88-server
Restart=always
RestartSec=5
Environment="DOWNLOAD_TOKEN=tok123"

[Install]
WantedBy=multi-user.target
EOF
    
    sudo mv /tmp/qh88.service /etc/systemd/system/
    sudo systemctl daemon-reload
    sudo systemctl enable qh88
    sudo systemctl restart qh88
    echo "✅ Systemd service đã được khởi động!"
fi

# Cấu hình Nginx nếu có file nginx.conf
if [ -f "nginx.conf" ] && [ -f "setup-domain.sh" ]; then
    echo "🌐 Cấu hình Nginx cho domain..."
    chmod +x setup-domain.sh
    sudo bash setup-domain.sh 2>/dev/null || echo "⚠️  Cần chạy thủ công: sudo bash setup-domain.sh"
fi

ENDSSH

echo ""
echo "✅ Deploy hoàn tất!"
echo ""
echo "📋 Thông tin VPS:"
echo "   - IP: $VPS_IP"
echo "   - Domain: $DOMAIN"
echo "   - Path: $VPS_PATH"
echo ""
echo "🔍 Kiểm tra ứng dụng:"
echo "   - http://$VPS_IP:8080"
echo "   - http://$DOMAIN (sau khi trỏ DNS)"
echo ""
echo "📝 Các bước tiếp theo:"
echo "   1. Trỏ DNS của $DOMAIN về IP: $VPS_IP"
echo "   2. SSH vào VPS: ssh $VPS_USER@$VPS_IP"
echo "   3. Cài SSL: sudo certbot --nginx -d $DOMAIN -d www.$DOMAIN"
echo ""
echo "📊 Xem logs:"
echo "   - Docker: ssh $VPS_USER@$VPS_IP 'cd $VPS_PATH && docker-compose logs -f'"
echo "   - Systemd: ssh $VPS_USER@$VPS_IP 'sudo journalctl -u qh88 -f'"

