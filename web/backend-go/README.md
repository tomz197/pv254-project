# backend-go

A minimal Go rewrite of the Python backend using net/http.

## Run

```bash
PORT=8000 FRONTEND_URL=http://localhost:5173 go run ./...
```

## Endpoints

- GET /health
- GET /models
- GET /course/{code}
- POST /recommendations
- POST /log_recommendation_feedback
- POST /log_user_feedback