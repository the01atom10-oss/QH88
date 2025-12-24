#!/bin/bash

# Script cấu hình domain qh88h1.com cho QH88
# Chạy với quyền root: sudo bash setup-domain.sh

DOMAIN="qh88h1.com"
APP_PORT=8080
VPS_IP="72.62.120.215"

echo "🌐 Cấu hình domain $DOMAIN cho QH88..."

# Kiểm tra quyền root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Vui lòng chạy với quyền root: sudo bash setup-domain.sh"
    exit 1
fi

# Cài đặt Nginx nếu chưa có
if ! command -v nginx &> /dev/null; then
    echo "📦 Cài đặt Nginx..."
    apt update
    apt install -y nginx
fi

# Copy cấu hình Nginx
echo "📝 Tạo cấu hình Nginx..."
cp nginx.conf /etc/nginx/sites-available/$DOMAIN

# Tạo symlink
if [ -f "/etc/nginx/sites-enabled/$DOMAIN" ]; then
    rm /etc/nginx/sites-enabled/$DOMAIN
fi
ln -s /etc/nginx/sites-available/$DOMAIN /etc/nginx/sites-enabled/

# Xóa default site nếu có
if [ -f "/etc/nginx/sites-enabled/default" ]; then
    rm /etc/nginx/sites-enabled/default
fi

# Test cấu hình
echo "🔍 Kiểm tra cấu hình Nginx..."
nginx -t

if [ $? -eq 0 ]; then
    echo "✅ Cấu hình hợp lệ!"
    systemctl reload nginx
    echo "✅ Nginx đã được reload!"
else
    echo "❌ Cấu hình có lỗi, vui lòng kiểm tra lại!"
    exit 1
fi

# Mở firewall
echo "🔥 Cấu hình firewall..."
if command -v ufw &> /dev/null; then
    ufw allow 'Nginx Full'
    ufw allow $APP_PORT/tcp
    echo "✅ Firewall đã được cấu hình!"
fi

echo ""
echo "✅ Hoàn tất cấu hình domain!"
echo ""
echo "📋 Các bước tiếp theo:"
echo "1. Trỏ DNS của domain $DOMAIN về IP VPS:"
echo "   - A record: $DOMAIN -> $VPS_IP"
echo "   - A record: www.$DOMAIN -> $VPS_IP"
echo ""
echo "2. Sau khi DNS đã trỏ xong (có thể mất vài phút đến vài giờ), cài SSL:"
echo "   sudo apt install certbot python3-certbot-nginx"
echo "   sudo certbot --nginx -d $DOMAIN -d www.$DOMAIN"
echo ""
echo "3. Sau khi cài SSL, mở file /etc/nginx/sites-available/$DOMAIN"
echo "   và bỏ comment các dòng HTTPS (dòng có # ở đầu)"
echo ""
echo "4. Reload Nginx:"
echo "   sudo nginx -t && sudo systemctl reload nginx"
echo ""
echo "🌐 Truy cập: http://$DOMAIN (sau khi cài SSL sẽ là https://$DOMAIN)"

