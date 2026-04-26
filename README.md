# Hypertube

Hypertube is a full-stack web application that allows users to search, discover, and stream movies directly from the browser.  
The platform supports real-time torrent-based video streaming using **HLS**, secure authentication, and a modern responsive interface.

---

## 🎥 Demo

<a href="https://www.youtube.com/watch?v=L0X-5HPQMko" target="_blank">
  <img src="https://img.shields.io/badge/Watch%20Full%20Demo-YouTube-red?logo=youtube&style=for-the-badge"/>
</a>

---

### HLS Transcoding & Streaming

> Torrent is resolved, converted to HLS segments on the fly, and streamed before the full download completes.

<!-- <video src="assets/demos/01_hls_transcoding.mp4" autoplay loop muted playsinline width="720"></video> -->


https://github.com/user-attachments/assets/138dbe96-f288-4ec5-a676-cedc5eb0a4ef


---

### Video Player — Quality Selection & Subtitles

> Switch between Auto / 480p / 720p / 1080p mid-playback. Subtitles are auto-loaded per language preference.



https://github.com/user-attachments/assets/29cd64f5-b613-4c69-a5aa-15d3466c9f4e

<!-- <video src="assets/demos/02_video_playback.mp4" autoplay loop muted playsinline width="720"></video> -->

---

### Discover & Search

> Browse the hero carousel, scroll the movie grid, or use live search with instant dropdown results.

<!-- <video src="assets/demos/06_discover_page.mp4" autoplay loop muted playsinline width="720"></video> -->
<!-- <video src="assets/demos/07_search.mp4" autoplay loop muted playsinline width="720"></video> -->




https://github.com/user-attachments/assets/3bef4598-8115-4ea7-b495-db1b10d0e650

https://github.com/user-attachments/assets/3983878e-9946-47cf-9148-c28c0ec24275


---

### Settings — Profile & Preferences

> Update personal info, swap avatar, change preferred language, and reset password.

<!-- <video src="assets/demos/04_settings_profile.mp4" autoplay loop muted playsinline width="720"></video> -->

https://github.com/user-attachments/assets/86a7ec32-02b6-4777-96a6-a4eaf1f53bac


---

## 🧱 Tech Stack

### Frontend
- **Next.js**
- **TypeScript**
- **React**
- **HLS.js** (HTTP Live Streaming)
- **WebSockets** (real-time updates)
- Responsive UI (Desktop & Mobile)

### Backend
- **Go (Golang)**
- RESTful API
- OAuth2 Authentication
- **HLS stream generation**
- **WebSocket server**

### Database
- **PostgreSQL**

### Infrastructure & DevOps
- **Docker & Docker Compose**
- **Nginx** (Reverse Proxy & HLS delivery)
- Environment-based configuration (`.env`)

---

## ✨ Features

### Authentication & Users
- User registration with:
  - Email
  - Username
  - First name
  - Last name
  - Password (securely hashed)
- OAuth authentication:
  - **42**
  - **Google**
  - **GitHub**
- Login / Logout
- Password reset via email
- User profile management (avatar, personal information)
- Public user profiles (email remains private)
- Preferred language selection (default: English)

---

### Library & Search
- Accessible only to authenticated users
- Movie search using **at least two external video sources**
- Display results as thumbnails
- Display popular movies when no search is performed
- Infinite scroll pagination (asynchronous loading)
- Sorting & filtering:
  - Name
  - Genre
  - IMDb rating
  - Production year
- Watched vs unwatched movie distinction

---

### 🎬 Video Streaming (HLS)
- Integrated video player
- Server-side torrent handling
- **Progressive streaming via HLS**
- Playback starts before full download
- Automatic HLS segment generation (`.m3u8`)
- Browser-compatible streaming
- On-the-fly video format conversion when required (MKV support)
- Cached movies stored on server
- Automatic cleanup of unwatched movies after 30 days

---

### 📝 Comments
- Authenticated users can post comments on movies
- Real-time comment updates via **WebSockets**
- Edit and delete own comments
- Display comment history per movie

---

### 🔌 Real-Time Features (WebSockets)
- Live streaming status updates
- Torrent download progress
- Real-time comment updates
- Non-blocking background processing

---

### 🔗 REST API
- OAuth2-secured REST API
- Public access to front-page movies
- Authenticated access to:
  - Users
  - Movies
  - Comments
- Strict HTTP method handling
- Fully RESTful architecture

### 🔗 PROJECT STRUCTURE
├── frontend/              # Next.js application
│   ├── src/
│   ├── public/
│   └── package.json
│
├── backend/               # Go API & streaming server
│   ├── cmd/
│   ├── internal/
│   ├── streaming/         # Torrent + HLS logic
│   ├── websocket/         # WebSocket handlers
│   └── main.go
│
├── nginx/
│   └── nginx.conf         # HLS & reverse proxy config
│
├── docker-compose.yml
├── .env.example
└── README.md


### 🔗 PROJECT STRUCTURE
- HLS streaming
- Multiple OAuth providers
- Web Sockets
- Background processing optimizations









