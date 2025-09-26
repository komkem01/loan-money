# Loan Money API

ระบบจัดการเงินกู้ พัฒนาด้วย Go และ PostgreSQL

## คุณสมบัติ

- ระบบสมัครสมาชิกและเข้าสู่ระบบ (Authentication)
- การเข้ารหัสรหัสผ่านด้วย Argon2id
- JWT Token สำหรับการยืนยันตัวตน
- การจัดการฐานข้อมูล PostgreSQL
- RESTful API

## โครงสร้างโปรเจ็ค

```
loan-money/
├── cmd/api/                 # Entry point ของแอปพลิเคชัน
│   └── main.go
├── internal/                # Internal packages (ไม่สามารถ import จากภายนอกได้)
│   ├── auth/               # Authentication middleware
│   ├── database/           # Database connection และ migration
│   ├── handlers/           # HTTP request handlers
│   └── models/            # Data models และ structures
├── pkg/utils/             # Utility functions (สามารถ import ได้)
├── .env.example          # ตัวอย่างไฟล์ environment variables
├── go.mod               # Go module dependencies
└── README.md           # เอกสารนี้
```

## การติดตั้ง

1. Clone repository
```bash
git clone <repository-url>
cd loan-money
```

2. ติดตั้ง dependencies
```bash
go mod tidy
```

3. สร้างไฟล์ .env จาก .env.example
```bash
cp .env.example .env
```

4. แก้ไขค่าในไฟล์ .env ให้เหมาะสม (รองรับ Supabase PostgreSQL)

5. รันแอปพลิเคชัน backend
```bash
go run cmd/api/main.go cmd/api/env.go
# หรือ
go build -o bin/loan-money.exe cmd/api/main.go cmd/api/env.go
./bin/loan-money.exe
```

6. รันแอปพลิเคชัน frontend (ในเทอร์มินัลใหม่)
```bash
python serve.py
```

7. เข้าใช้งาน
- Backend API: http://localhost:8080
- Frontend: http://localhost:3000

## API Endpoints

### Authentication

#### Register User
```http
POST /api/v1/register
Content-Type: application/json

{
    "username": "john_doe",
    "password": "securepassword123",
    "full_name": "John Doe"
}

Response:
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
        "id": "uuid",
        "username": "john_doe",
        "full_name": "John Doe",
        "created_at": "2025-09-27T10:30:00Z"
    }
}
```

#### Login
```http
POST /api/v1/login
Content-Type: application/json

{
    "username": "john_doe", 
    "password": "securepassword123"
}

Response:
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
        "id": "uuid",
        "username": "john_doe",
        "full_name": "John Doe",
        "created_at": "2025-09-27T10:30:00Z"
    }
}
```

#### Get Profile (Protected)
```http
GET /api/v1/profile
Authorization: Bearer <jwt_token>

Response:
{
    "user_id": "uuid",
    "username": "john_doe"
}
```

### Health Check
```http
GET /health

Response:
{
    "status": "healthy"
}
```

## Frontend Pages

### หน้าเข้าสู่ระบบ
- URL: http://localhost:3000/index.html
- ฟีเจอร์: เข้าสู่ระบบด้วย username/password
- เชื่อมต่อกับ `/api/v1/login`

### หน้าสมัครสมาชิก  
- URL: http://localhost:3000/register.html
- ฟีเจอร์: สมัครสมาชิกใหม่
- เชื่อมต่อกับ `/api/v1/register`

### หน้าทดสอบ API
- URL: http://localhost:3000/test.html
- ฟีเจอร์: ทดสอบ API endpoints
- แสดงสถานะ authentication

### หน้า Dashboard
- URL: http://localhost:3000/dashboard.html
- ฟีเจอร์: หน้าหลักของระบบ (ต้อง login)

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR NOT NULL UNIQUE,
    password VARCHAR NOT NULL,
    full_name VARCHAR,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Loans Table
```sql
CREATE TABLE loans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    borrower_name VARCHAR NOT NULL,
    amount NUMERIC NOT NULL,
    status VARCHAR NOT NULL DEFAULT 'active',
    loan_date DATE NOT NULL,
    due_date DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Transactions Table
```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    loan_id UUID NOT NULL REFERENCES loans(id),
    amount NUMERIC NOT NULL,
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## ความปลอดภัย

- รหัสผ่านถูกเข้ารหัสด้วย Argon2id
- ใช้ JWT สำหรับ authentication
- Middleware สำหรับตรวจสอบสิทธิ์การเข้าถึง
- CORS configuration สำหรับ cross-origin requests

## Dependencies

- `github.com/gorilla/mux` - HTTP router
- `github.com/lib/pq` - PostgreSQL driver  
- `github.com/golang-jwt/jwt/v5` - JWT implementation
- `github.com/google/uuid` - UUID generation
- `golang.org/x/crypto` - Cryptography (Argon2id)
- `github.com/rs/cors` - CORS middleware

## HTTP Methods ที่ใช้

- `GET` - สำหรับดึงข้อมูล (ดู profile, ดูรายการ loans)
- `POST` - สำหรับสร้างข้อมูลใหม่ (register, login, สร้าง loan)
- `PATCH` - สำหรับอัปเดตข้อมูลบางส่วน (อัปเดต profile, สถานะ loan)
- `DELETE` - สำหรับลบข้อมูล (ลบ loan, ลบ transaction)

## การพัฒนาเพิ่มเติม

- [ ] CRUD operations สำหรับ loans
- [ ] CRUD operations สำหรับ transactions  
- [ ] Dashboard statistics
- [ ] User management
- [ ] Audit logging
- [ ] API rate limiting
- [ ] Input validation middleware
- [ ] Database migration system