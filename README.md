# 🎬 Hypertube

A web app for the 21st century: search, stream, and watch videos directly from torrents.

---

## 🚀 Project Overview

Hypertube is a full-stack web application built with **Spring Boot** (backend) and **Angular** (frontend).  
It enables users to **search for movies**, **stream them instantly**, and manage their profiles, all while respecting strict security and UI/UX best practices.

---

## 📚 Features

### 🔑 Authentication & User Management
- Register with **email, username, first/last name, password** (securely hashed).
- Login via:
  - Username & password
  - **Omniauth** (42 + one other provider of choice)
- Password reset via email.
- Logout from any page with one click.
- Multi-language support (default **English**).
- User profile management (update email, profile picture, personal info).
- Public user profiles (without exposing email).

### 🎥 Library
- Accessible only to authenticated users.
- **Search movies** via at least **two external APIs** (OMDb/TMDb + another source).
- Display movies as thumbnails:
  - Title, year, IMDb rating, cover image.
  - Watched/unwatched differentiation.
  - Infinite scrolling (auto-load next page).
  - Sort & filter by genre, rating, year, etc.
- Show trending/popular movies if no search is performed.

### ▶️ Video Player
- Movie details: summary, cast, year, length, rating, cover.
- Built-in **video player** with:
  - Subtitles (auto-download English + user’s preferred language if available).
  - On-the-fly format conversion (support at least `mkv` → mp4/webm).
- Background torrent download & streaming:
  - Start playback once enough data is available.
  - Movies cached on server (deleted if unwatched for 30+ days).
- Comments section:
  - Post/view comments.
  - Display author, date, and content.

### 🔌 RESTful API (OAuth2-secured)
- `POST /oauth/token` → Get auth token.
- `GET /users`, `GET /users/:id`, `PATCH /users/:id` → Manage profiles.
- `GET /movies`, `GET /movies/:id` → Retrieve movie info.
- `GET /comments`, `GET /comments/:id` → Fetch comments.
- `POST /comments` or `POST /movies/:id/comments` → Post comment.
- Proper **HTTP status codes** for invalid routes/actions.

---

## 🛠️ Tech Stack

- **Backend:** Spring Boot (Java), OAuth2, RESTful API
- **Frontend:** Angular + [UI library of choice] (e.g. Angular Material / NG-Zorro / PrimeNG)
- **Database:** PostgreSQL / MySQL
- **Torrent Handling:** Custom server-side implementation (⚠️ no WebTorrent/Peerflix/Pulsar allowed)
- **APIs:** OMDb / TMDb + another torrent/video source
- **Deployment:** Docker, Nginx/Apache

---

## 📱 UI/UX Requirements

- Responsive design (desktop + mobile).
- Minimum layout: **Header, Main, Footer**.
- Form validation & input sanitization.
- Secure handling of uploads and user input.

---

## ⚖️ Rules & Constraints

- No plain-text passwords (must be hashed/salted).
- No SQL injections or XSS vulnerabilities.
- No storing credentials in Git — use `.env` (excluded from repo).
- Application must work on **latest Firefox & Chrome**.
- No errors/warnings in client/server console.

---

## 🎁 Bonus Features (optional)
- Extra Omniauth providers (Google, GitHub, etc.)
- Multiple video resolutions.
- Streaming via MediaStream API.
- Extended API endpoints (e.g., add/remove movies).

---

## 📦 Installation & Setup

### 1️⃣ Backend
```bash
cd backend
cp .env.example .env   # configure DB + API keys
./mvnw spring-boot:run
```

### 2️⃣ Frontend
```bash
cd frontend
npm install
ng serve
```

### 3️⃣ Access
- Frontend: http://localhost:4200
- Backend API: http://localhost:8080

---

## 👨‍👩‍👧 Team
This project is built by **4 teammates** as part of the Hypertube assignment.

- Backend: Spring Boot developers
- Frontend: Angular developers
- UI/UX: modern, responsive, mobile-first design

---

## 📜 License
This project is for **educational purposes only**.
