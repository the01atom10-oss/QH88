# 🚀 Hướng dẫn Deploy Nhanh QH88 với Domain qh88h1.com

## Bước 1: Upload code lên VPS

```bash
# Từ máy local
scp -r QH88/ user@your-vps-ip:/opt/qh88
```

## Bước 2: SSH vào VPS và deploy

```bash
ssh user@your-vps-ip
cd /opt/qh88
```

## Bước 3: Chạy ứng dụng (Chọn 1 trong 2 cách)

### Cách A: Docker (Khuyến nghị)

```bash
# Tạo file .env
echo "DOWNLOAD_TOKEN=tok123" > .env

# Chạy với Docker
docker-compose up -d --build

# Kiểm tra
docker-compose logs -f
```

### Cách B: Binary trực tiếp

```bash
# Build binary (nếu chưa build)
bash build.sh

# Hoặc build trên máy local và upload
# scp qh88-server user@vps:/opt/qh88/

# Chạy
chmod +x qh88-server
./qh88-server
```

## Bước 4: Cấu hình Domain qh88h1.com

```bash
# Chạy script tự động
chmod +x setup-domain.sh
sudo bash setup-domain.sh
```

Script sẽ:
- ✅ Cài Nginx (nếu chưa có)
- ✅ Tạo cấu hình cho qh88h1.com
- ✅ Cấu hình firewall
- ✅ Hiển thị IP VPS để trỏ DNS

## Bước 5: Trỏ DNS

Tại nhà cung cấp domain (Namecheap, GoDaddy, v.v.):

```
A Record:
- Name: @ hoặc qh88h1.com
- Value: [IP VPS hiển thị trong script]

A Record:
- Name: www
- Value: [IP VPS]
```

Đợi DNS propagate (5 phút - 24 giờ)

## Bước 6: Cài SSL/HTTPS

```bash
# Cài Certbot
sudo apt install -y certbot python3-certbot-nginx

# Cài SSL tự động
sudo certbot --nginx -d qh88h1.com -d www.qh88h1.com

# Certbot sẽ tự động:
# - Tạo SSL certificate
# - Cấu hình HTTPS
# - Redirect HTTP -> HTTPS
```

## Bước 7: Kiểm tra

- ✅ https://qh88h1.com
- ✅ https://www.qh88h1.com
- ✅ https://qh88h1.com/download?token=tok123

## Quản lý Service

### Nếu dùng Docker:
```bash
docker-compose restart    # Khởi động lại
docker-compose logs -f    # Xem logs
docker-compose down       # Dừng
```

### Nếu dùng Binary:
```bash
# Tạo systemd service
sudo cp qh88.service /etc/systemd/system/
sudo nano /etc/systemd/system/qh88.service  # Sửa user và token

sudo systemctl daemon-reload
sudo systemctl enable qh88
sudo systemctl start qh88
sudo systemctl status qh88
```

## Troubleshooting

**Domain không truy cập được?**
```bash
# Kiểm tra DNS
dig qh88h1.com
nslookup qh88h1.com

# Kiểm tra Nginx
sudo nginx -t
sudo systemctl status nginx

# Kiểm tra ứng dụng
curl http://localhost:8080
```

**SSL không hoạt động?**
```bash
# Kiểm tra certificate
sudo certbot certificates

# Renew SSL (tự động renew mỗi 90 ngày)
sudo certbot renew
```

**Xem logs:**
```bash
# Nginx logs
sudo tail -f /var/log/nginx/qh88h1.com.error.log

# App logs (Docker)
docker-compose logs -f

# App logs (Systemd)
sudo journalctl -u qh88 -f
```

