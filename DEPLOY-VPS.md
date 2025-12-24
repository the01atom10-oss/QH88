# 🚀 Hướng dẫn Deploy QH88 lên VPS 72.62.120.215

## Thông tin VPS
- **IP**: 72.62.120.215
- **User**: root
- **Domain**: qh88h1.com
- **Path**: /opt/qh88

---

## Phương pháp 1: Deploy Tự Động (Khuyến nghị)

### Bước 1: Chạy script deploy tự động

```bash
cd QH88
chmod +x deploy-to-vps.sh
bash deploy-to-vps.sh
```

Script sẽ tự động:
- ✅ Build binary cho Linux
- ✅ Upload code lên VPS
- ✅ Cấu hình và khởi động service
- ✅ Cấu hình Nginx (nếu có)

---

## Phương pháp 2: Deploy Thủ Công

### Bước 1: Upload code lên VPS

```bash
# Từ máy local
cd QH88
scp -r * root@72.62.120.215:/opt/qh88
```

### Bước 2: SSH vào VPS

```bash
ssh root@72.62.120.215
cd /opt/qh88
```

### Bước 3: Chạy ứng dụng

#### Cách A: Sử dụng Docker (Khuyến nghị)

```bash
# Tạo file .env
echo "DOWNLOAD_TOKEN=tok123" > .env

# Build và chạy
docker-compose up -d --build

# Kiểm tra
docker-compose logs -f
```

#### Cách B: Build và chạy binary trực tiếp

```bash
# Cài đặt Go (nếu chưa có)
apt update
apt install -y golang-go

# Build binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o qh88-server main.go

# Chạy
chmod +x qh88-server
./qh88-server
```

#### Cách C: Tạo systemd service (Tự động khởi động)

```bash
# Tạo service file
cat > /etc/systemd/system/qh88.service << 'EOF'
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

# Khởi động service
systemctl daemon-reload
systemctl enable qh88
systemctl start qh88
systemctl status qh88
```

### Bước 4: Cấu hình Domain qh88h1.com

```bash
# Chạy script tự động
chmod +x setup-domain.sh
bash setup-domain.sh
```

Hoặc cấu hình thủ công:

```bash
# Cài Nginx
apt update
apt install -y nginx

# Copy cấu hình
cp nginx.conf /etc/nginx/sites-available/qh88h1.com
ln -s /etc/nginx/sites-available/qh88h1.com /etc/nginx/sites-enabled/
rm /etc/nginx/sites-enabled/default

# Test và reload
nginx -t
systemctl reload nginx

# Mở firewall
ufw allow 'Nginx Full'
ufw allow 8080/tcp
```

### Bước 5: Trỏ DNS

Tại nhà cung cấp domain, thêm các record:

```
Type: A
Name: @ (hoặc qh88h1.com)
Value: 72.62.120.215

Type: A
Name: www
Value: 72.62.120.215
```

Đợi DNS propagate (5 phút - 24 giờ)

### Bước 6: Cài SSL/HTTPS

```bash
# Cài Certbot
apt install -y certbot python3-certbot-nginx

# Cài SSL tự động
certbot --nginx -d qh88h1.com -d www.qh88h1.com

# Certbot sẽ tự động cấu hình HTTPS
```

### Bước 7: Kiểm tra

- ✅ http://72.62.120.215:8080
- ✅ http://qh88h1.com (sau khi DNS trỏ xong)
- ✅ https://qh88h1.com (sau khi cài SSL)
- ✅ https://qh88h1.com/download?token=tok123

---

## Quản lý Service

### Nếu dùng Docker:

```bash
ssh root@72.62.120.215
cd /opt/qh88

docker-compose restart    # Khởi động lại
docker-compose logs -f    # Xem logs
docker-compose down       # Dừng
docker-compose up -d      # Khởi động
```

### Nếu dùng Systemd:

```bash
ssh root@72.62.120.215

systemctl restart qh88    # Khởi động lại
systemctl status qh88      # Xem trạng thái
journalctl -u qh88 -f     # Xem logs
systemctl stop qh88       # Dừng
systemctl start qh88       # Khởi động
```

---

## Kiểm tra và Troubleshooting

### Kiểm tra ứng dụng đang chạy:

```bash
ssh root@72.62.120.215

# Kiểm tra port 8080
netstat -tulpn | grep 8080
# Hoặc
ss -tulpn | grep 8080

# Test local
curl http://localhost:8080
```

### Kiểm tra Nginx:

```bash
ssh root@72.62.120.215

# Test cấu hình
nginx -t

# Xem logs
tail -f /var/log/nginx/qh88h1.com.error.log
tail -f /var/log/nginx/qh88h1.com.access.log

# Kiểm tra status
systemctl status nginx
```

### Kiểm tra DNS:

```bash
# Từ máy local
dig qh88h1.com
nslookup qh88h1.com
ping qh88h1.com
```

### Kiểm tra SSL:

```bash
ssh root@72.62.120.215

# Xem certificates
certbot certificates

# Renew SSL (tự động mỗi 90 ngày)
certbot renew
```

### Xem logs ứng dụng:

```bash
ssh root@72.62.120.215

# Docker
cd /opt/qh88 && docker-compose logs -f

# Systemd
journalctl -u qh88 -f

# Binary trực tiếp
# Xem output console hoặc redirect vào file
```

---

## Backup dữ liệu

```bash
ssh root@72.62.120.215

# Backup file logins.json
cd /opt/qh88
cp data/logins.json data/logins.json.backup.$(date +%Y%m%d_%H%M%S)

# Hoặc backup toàn bộ thư mục data
tar -czf backup-qh88-$(date +%Y%m%d).tar.gz data/
```

---

## Cập nhật code mới

```bash
# Từ máy local
cd QH88
# Sửa code...

# Upload lại
scp -r * root@72.62.120.215:/opt/qh88

# SSH vào VPS và restart
ssh root@72.62.120.215
cd /opt/qh88

# Nếu dùng Docker
docker-compose down
docker-compose up -d --build

# Nếu dùng Systemd
systemctl restart qh88
```

---

## Firewall

```bash
ssh root@72.62.120.215

# Kiểm tra firewall
ufw status

# Mở các port cần thiết
ufw allow 22/tcp      # SSH
ufw allow 80/tcp      # HTTP
ufw allow 443/tcp     # HTTPS
ufw allow 8080/tcp    # App port
ufw allow 'Nginx Full' # Hoặc dùng lệnh này

# Enable firewall (nếu chưa)
ufw enable
```

---

## Thông tin liên hệ

- **VPS IP**: 72.62.120.215
- **SSH**: `ssh root@72.62.120.215`
- **Domain**: qh88h1.com
- **App Port**: 8080
- **Download Token**: tok123 (có thể đổi trong file .env)

