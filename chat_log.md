# Hypertube Movie Download Chat Log

## Summary

Successfully implemented and tested movie search and download functionality in Hypertube application.

## Key Commands Discovered

### Authentication

```bash
# Register user
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your_username",
    "email": "your_email@example.com",
    "password": "your_password",
    "FirstName": "Your",
    "LastName": "Name"
  }'

# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your_email@example.com",
    "password": "your_password"
  }'
```

### Movie Operations

```bash
# Search torrents
curl "http://localhost:8080/torrents/search?title=avengers" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Start download
curl -X POST "http://localhost:8080/torrents/download" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "movie_id": 99999,
    "magnet": "magnet:?xt=urn:btih:3B5B9C573BC71221E6A5F336D4D32CFEB2A20A36&dn=Lego+Marvel+Avengers%3A+Mission+Demolition",
    "quality": "2160p"
  }'

# Check progress
curl "http://localhost:8080/torrents/progress?movie_id=99999&quality=2160p" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Stream movie
curl "http://localhost:8080/stream/99999?quality=2160p" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Results Achieved

- ✅ User registration and login working
- ✅ Movie search returning torrents with seeders/leechers info
- ✅ Download started successfully
- ✅ Progress monitoring working (showing ~0.40% downloaded)
- ✅ Stream ready status achieved
- ✅ Repository pushed to 'ayuup' branch

## Notes

- Torrents with more seeders (42 vs 8) start downloading much faster
- Progressive streaming allows watching while downloading
- JWT tokens expire and need renewal for continued API access
- API expects specific field names: movie_id (int), magnet (string), quality (string)
