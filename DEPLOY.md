# Hướng dẫn Deploy QH88 lên VPS Linux

## Phương pháp 1: Sử dụng Docker (Khuyến nghị)

### Yêu cầu
- VPS Linux (Ubuntu/Debian/CentOS)
- Docker và Docker Compose đã cài đặt

### Các bước deploy

1. **Upload code lên VPS**
```bash
# Trên máy local, upload thư mục QH88 lên VPS
scp -r QH88/ user@your-vps-ip:/opt/qh88
```

2. **SSH vào VPS**
```bash
ssh user@your-vps-ip
cd /opt/qh88
```

3. **Tạo file .env**
```bash
cp .env.example .env
nano .env  # Chỉnh sửa DOWNLOAD_TOKEN nếu cần
```

4. **Build và chạy với Docker Compose**
```bash
docker-compose up -d --build
```

5. **Kiểm tra logs**
```bash
docker-compose logs -f
```

6. **Kiểm tra ứng dụng**
- Mở trình duyệt: `http://your-vps-ip:8080`
- API download: `http://your-vps-ip:8080/download?token=YOUR_TOKEN`

### Quản lý service

```bash
# Dừng service
docker-compose down

# Khởi động lại
docker-compose restart

# Xem logs
docker-compose logs -f

# Cập nhật code mới
git pull  # hoặc upload code mới
docker-compose up -d --build
```

---

## Phương pháp 2: Build binary và chạy trực tiếp

### Yêu cầu
- VPS Linux với Go đã cài đặt (hoặc build trên máy local)

### Các bước deploy

1. **Build binary trên máy local (Windows)**
```bash
cd QH88
# Sử dụng WSL hoặc Git Bash
bash build.sh
```

2. **Upload binary và files lên VPS**
```bash
scp qh88-server user@your-vps-ip:/opt/qh88/
scp -r web/ user@your-vps-ip:/opt/qh88/
scp -r data/ user@your-vps-ip:/opt/qh88/
scp .env user@your-vps-ip:/opt/qh88/
```

3. **SSH vào VPS và setup**
```bash
ssh user@your-vps-ip
cd /opt/qh88
chmod +x qh88-server
mkdir -p data
```

4. **Tạo systemd service (tự động khởi động)**
```bash
sudo nano /etc/systemd/system/qh88.service
```

Nội dung file:
```ini
[Unit]
Description=QH88 Server
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/opt/qh88
ExecStart=/opt/qh88/qh88-server
Restart=always
RestartSec=5
Environment="DOWNLOAD_TOKEN=your-token-here"

[Install]
WantedBy=multi-user.target
```

5. **Khởi động service**
```bash
sudo systemctl daemon-reload
sudo systemctl enable qh88
sudo systemctl start qh88
sudo systemctl status qh88
```

6. **Xem logs**
```bash
sudo journalctl -u qh88 -f
```

---

## Phương pháp 3: Cấu hình Domain qh88h1.com với Nginx

### Cấu hình domain qh88h1.com (Khuyến nghị)

1. **Upload file cấu hình lên VPS**
```bash
scp QH88/nginx.conf user@vps-ip:/opt/qh88/
scp QH88/setup-domain.sh user@vps-ip:/opt/qh88/
```

2. **SSH vào VPS và chạy script tự động**
```bash
ssh user@vps-ip
cd /opt/qh88
chmod +x setup-domain.sh
sudo bash setup-domain.sh
```

Script sẽ tự động:
- Cài đặt Nginx (nếu chưa có)
- Tạo cấu hình cho domain qh88h1.com
- Cấu hình firewall
- Hiển thị hướng dẫn các bước tiếp theo

3. **Trỏ DNS về VPS**
Tại nhà cung cấp domain, thêm các record:
```
Type: A
Name: @ (hoặc qh88h1.com)
Value: [IP của VPS]

Type: A  
Name: www
Value: [IP của VPS]
```

4. **Cài đặt SSL/HTTPS (Sau khi DNS đã trỏ xong)**
```bash
# Cài đặt Certbot
sudo apt update
sudo apt install -y certbot python3-certbot-nginx

# Cài SSL tự động
sudo certbot --nginx -d qh88h1.com -d www.qh88h1.com

# Certbot sẽ tự động cấu hình HTTPS và redirect HTTP -> HTTPS
```

5. **Kích hoạt HTTPS trong Nginx config**
Sau khi cài SSL, mở file cấu hình:
```bash
sudo nano /etc/nginx/sites-available/qh88h1.com
```

Bỏ comment (xóa dấu #) ở phần cấu hình HTTPS, và comment phần HTTP redirect:
```nginx
server {
    listen 80;
    server_name qh88h1.com www.qh88h1.com;
    return 301 https://$server_name$request_uri;  # Bỏ comment dòng này
}

server {
    listen 443 ssl http2;
    server_name qh88h1.com www.qh88h1.com;
    # Bỏ comment toàn bộ block này
    ...
}
```

Reload Nginx:
```bash
sudo nginx -t && sudo systemctl reload nginx
```

6. **Kiểm tra**
- HTTP: `http://qh88h1.com` (sẽ redirect sang HTTPS)
- HTTPS: `https://qh88h1.com`
- API: `https://qh88h1.com/download?token=YOUR_TOKEN`

### Cấu hình thủ công (Nếu không dùng script)

1. **Cài đặt Nginx**
```bash
sudo apt update
sudo apt install nginx
```

2. **Copy file cấu hình**
```bash
sudo cp nginx.conf /etc/nginx/sites-available/qh88h1.com
sudo ln -s /etc/nginx/sites-available/qh88h1.com /etc/nginx/sites-enabled/
sudo rm /etc/nginx/sites-enabled/default  # Xóa default site
```

3. **Test và reload**
```bash
sudo nginx -t
sudo systemctl reload nginx
```

---

## Firewall

Mở port cần thiết:
```bash
# Ubuntu/Debian
sudo ufw allow 8080/tcp    # Port ứng dụng
sudo ufw allow 80/tcp      # HTTP (nếu dùng Nginx)
sudo ufw allow 443/tcp     # HTTPS (nếu dùng SSL)
sudo ufw allow 'Nginx Full'  # Hoặc dùng lệnh này cho cả HTTP và HTTPS
sudo ufw reload

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --reload
```

---

## Backup dữ liệu

File dữ liệu quan trọng: `data/logins.json`

```bash
# Backup định kỳ
cp data/logins.json data/logins.json.backup.$(date +%Y%m%d)
```

---

## Troubleshooting

1. **Port đã được sử dụng**
```bash
# Kiểm tra port 8080
sudo netstat -tulpn | grep 8080
# Hoặc
sudo lsof -i :8080
```

2. **Permission denied**
```bash
chmod +x qh88-server
chmod 755 data/
```

3. **Xem logs**
- Docker: `docker-compose logs -f`
- Systemd: `sudo journalctl -u qh88 -f`
- Binary trực tiếp: kiểm tra output console

