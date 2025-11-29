# Phase 1a: Auth Service Implementation Plan

**Timeline**: Weeks 2–3  
**Goal**: Build JWT-based authentication with user registration/login endpoints  
**Acceptance Criteria**: 
- User can register with email, phone, name, and password
- User can login and receive JWT access token + refresh token
- Protected routes validate JWT tokens
- Passwords are securely hashed with bcrypt
- Role-based access control (student/teacher/admin) works

## Architecture

```
Browser/Client
    │
    ├─→ POST /api/v1/auth/register (email, password, name, phone, role)
    │   └─→ Validate input → Hash password → Store in DB → Return user ID
    │
    ├─→ POST /api/v1/auth/login (email, password)
    │   └─→ Verify password → Generate JWT → Return tokens
    │
    ├─→ GET /api/v1/protected (with Authorization: Bearer <token>)
    │   └─→ JWT middleware validates token → Allow/Deny
    │
    └─→ POST /api/v1/auth/refresh (with refresh_token)
        └─→ Validate refresh token → Issue new access token
```

## Implementation Components

### 1. JWT Token Service (`pkg/auth/token.go`)

**Responsibilities:**
- Create access tokens (24-hour expiry)
- Create refresh tokens (7-day expiry)
- Validate tokens and extract claims
- Handle token refresh logic

**Key Types:**
```go
type TokenClaims struct {
    UserID   string   `json:"user_id"`
    Email    string   `json:"email"`
    Role     string   `json:"role"`
    ExpiresAt int64   `json:"exp"`
}

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}
```

**Methods:**
- `NewTokenService(secret string)` - Initialize with JWT secret
- `GenerateTokenPair(userID, email, role string)` - Create both tokens
- `ValidateToken(token string)` - Parse & validate JWT
- `RefreshAccessToken(refreshToken string)` - Issue new access token

### 2. Password Service (`pkg/auth/password.go`)

**Responsibilities:**
- Hash passwords using bcrypt (cost 12)
- Verify passwords against hashes

**Methods:**
- `HashPassword(password string)` - Bcrypt hash
- `VerifyPassword(hash, password string)` - Compare hash & password

### 3. User Database Layer (`pkg/auth/user_store.go`)

**Responsibilities:**
- CRUD operations on `users` table
- Validate email uniqueness
- Handle concurrent registrations

**Methods:**
- `CreateUser(ctx context.Context, user *User)` - Insert user
- `GetUserByEmail(ctx context.Context, email string)` - Fetch by email
- `GetUserByID(ctx context.Context, id string)` - Fetch by ID
- `UpdateLastLogin(ctx context.Context, userID string)` - Update timestamp

### 4. Auth Endpoints (`pkg/auth/handlers.go`)

**Endpoints:**

#### `POST /api/v1/auth/register`
```json
Request:
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "full_name": "أحمد محمد",
  "phone": "+963123456789",
  "role": "student"
}

Response (201 Created):
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "full_name": "أحمد محمد",
  "role": "student"
}

Error (400):
{
  "error": "Email already exists"
}
```

#### `POST /api/v1/auth/login`
```json
Request:
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}

Response (200 OK):
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 86400,
  "user": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "student"
  }
}

Error (401):
{
  "error": "Invalid email or password"
}
```

#### `POST /api/v1/auth/refresh`
```json
Request:
{
  "refresh_token": "eyJhbGc..."
}

Response (200 OK):
{
  "access_token": "eyJhbGc...",
  "expires_in": 86400
}

Error (401):
{
  "error": "Refresh token expired or invalid"
}
```

#### `POST /api/v1/auth/logout`
```json
Request:
{
  "refresh_token": "eyJhbGc..."
}

Response (200 OK):
{
  "message": "Logged out successfully"
}
```

### 5. Middleware (`pkg/auth/middleware.go`)

**JWT Validation Middleware:**
- Extract token from `Authorization: Bearer <token>` header
- Validate signature and expiration
- Extract claims and add to request context
- Return 401 if invalid

**RBAC Middleware (optional for Phase 1a, required for Phase 1b):**
- Check user role against required roles
- Return 403 if unauthorized

**Example usage in route handler:**
```go
http.HandleFunc("/api/v1/protected", authMiddleware(protectedHandler))
```

### 6. Input Validation

**Register endpoint:**
- Email: valid format, not already registered
- Password: min 8 chars, at least one uppercase, one digit
- Full name: non-empty, max 255 chars
- Phone: valid Syrian format (optional)
- Role: must be "student", "teacher", or "admin"

**Login endpoint:**
- Email: required, valid format
- Password: required, non-empty

**Refresh endpoint:**
- Refresh token: required, valid format

## Database Integration

### Existing `users` Table (from Phase 0)
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    full_name VARCHAR(255),
    role VARCHAR(50) NOT NULL CHECK (role IN ('student', 'teacher', 'admin')),
    password_hash VARCHAR(255) NOT NULL,
    locale VARCHAR(10) DEFAULT 'ar_SY',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Optional: Add `refresh_tokens` Table (for token revocation)
```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

This allows revocation of specific refresh tokens (logout functionality).

## Error Handling

**Standard Error Response:**
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "status": 400
}
```

**Error Codes:**
- `INVALID_INPUT` - Validation failed (400)
- `EMAIL_EXISTS` - Email already registered (409)
- `INVALID_CREDENTIALS` - Wrong password (401)
- `TOKEN_EXPIRED` - JWT expired (401)
- `TOKEN_INVALID` - Malformed JWT (401)
- `ROLE_REQUIRED` - Insufficient permissions (403)

## Testing Strategy

### Unit Tests
- Token generation & validation
- Password hashing & verification
- Input validation for register/login
- Role-based access checks

### Integration Tests
- Register new user → verify in DB
- Login → receive valid tokens
- Use access token on protected route → succeeds
- Use expired token → 401
- Use refresh token → new access token issued
- Logout → refresh token invalidated

### Manual Testing
```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test1234!","full_name":"Test User","phone":"+963123456789","role":"student"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test1234!"}'

# Use token on protected route (once routes exist)
curl -X GET http://localhost:8080/api/v1/protected \
  -H "Authorization: Bearer <access_token>"
```

## Configuration Needed

Add to `.env`:
```
JWT_SECRET=your-very-long-secret-key-change-in-production
JWT_EXPIRY_HOURS=24
JWT_REFRESH_EXPIRY_HOURS=168
BCRYPT_COST=12
```

## Dependencies to Add

```bash
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
```

## Implementation Sequence

1. **Token Service** (`token.go`) - Core JWT logic
2. **Password Service** (`password.go`) - Bcrypt functions
3. **User Store** (`user_store.go`) - Database layer
4. **Handlers** (`handlers.go`) - API endpoints
5. **Middleware** (`middleware.go`) - JWT validation
6. **Main** (`cmd/main.go`) - Register routes
7. **Tests** - Unit & integration tests
8. **Documentation** - API spec

## Success Criteria for Phase 1a

- [ ] User registration works end-to-end
- [ ] User login returns valid JWT tokens
- [ ] Protected routes enforce JWT validation
- [ ] Passwords are hashed with bcrypt
- [ ] Role-based access control working
- [ ] Input validation prevents malicious input
- [ ] Error messages are clear and standard
- [ ] All code reviewed and approved
- [ ] Unit tests pass (>80% coverage)
- [ ] Manual smoke tests pass

## Files to Be Created

```
pkg/auth/
├── token.go           - JWT operations
├── password.go        - Bcrypt password handling
├── user_store.go      - User database operations
├── handlers.go        - HTTP endpoint handlers
├── middleware.go      - JWT validation middleware
├── types.go           - Shared types
└── auth_test.go       - Unit tests (if included)
```

## Next Phase Integration

Phase 1a outputs feed into:
- **Phase 1b**: Signalling endpoint will check JWT tokens
- **Phase 2c**: Chat will use authenticated user context
- **Phase 3**: Admin dashboard will use role checks

---

**Status**: Ready for implementation  
**Approval**: Awaiting your approval to proceed  

Should I begin coding Phase 1a now?
