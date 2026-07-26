# REST API Patterns

## ⚠️ Critical API File Rules

**‼️ EVERY `.api` file MUST follow these rules:**

### 1. MUST Set `go_package` Option

```api
info(
    title: "User API"
    desc: "User management API"
    author: "pzero"
    version: "v1"
    go_package: "user"  // ‼️ REQUIRED - MUST set go_package
)
```

### 2. MUST Set `group` and `compact_handler` in @server Block

```api
@server(
    prefix: /api/v1
    group: user  // ‼️ REQUIRED - MUST set group
    compact_handler: true      // ‼️ REQUIRED merge handler to one file
    middleware: Auth
)
service user-api {
    @handler Create  // ✅ No group prefix needed
    post /users (CreateRequest) returns (CreateResponse)
}
```

### 3. Benefits of Using `group`

When `group` is set in `@server`:

- ✅ **Handler names don't need group prefix**
  - ❌ Wrong: `@handler UserCreate`, `@handler UserGet`
  - ✅ Correct: `@handler Create`, `@handler Get`

- ✅ **Type names don't need group prefix**
  - ❌ Wrong: `UserCreateRequest`, `UserGetResponse`
  - ✅ Correct: `CreateRequest`, `GetResponse`

- ✅ **Cleaner, simpler naming** - No repetitive prefixes

### ❌ WRONG - Without group

```api
// No group set in @server
@server(
    prefix: /api/v1
)
service user-api {
    @handler CreateUser
    post /users (CreateUserRequest) returns (CreateUserResponse)
}
```

### ✅ CORRECT - With group

```api
// group set in @server
@server(
    prefix: /api/v1
    group: user  // ‼️ REQUIRED
    compact_handler: true      // ‼️ REQUIRED merge handler to one file
)
service user-api {
    @handler Create  // ✅ No group prefix needed
    post /users (CreateRequest) returns (CreateResponse)
}
```

### ❌ WRONG - Without compact_handler

```api
// No compact_handler set in @server
@server(
    prefix: /api/v1
)
service user-api {
    @handler CreateUser
    post /users (CreateUserRequest) returns (CreateUserResponse)
}
```

### ✅ CORRECT - With compact_handler

```api
// group set in @server
@server(
    prefix: /api/v1
    group: user  // ‼️ REQUIRED
    compact_handler: true      // ‼️ REQUIRED merge handler to one file
)
service user-api {
    @handler Create  // ✅ No group prefix needed
    post /users (CreateRequest) returns (CreateResponse)
}
```

---

## Core Architecture

### Three-Layer Pattern

pzero REST APIs follow a strict three-layer architecture:

1. **Handler Layer** (`internal/handler/`) - HTTP concerns only
2. **Logic Layer** (`internal/logic/`) - Business logic implementation
3. **Service Context** (`internal/svc/`) - Dependency injection

```
HTTP Request → Handler → Logic → External Services/Database
                  ↓
            Service Context (dependencies)
```

## Request/Response Types

### ✅ Correct Pattern

Define clear types with proper validation tags. **Note: With `group` set in `@server`, type names don't need group prefix:**

```go
// API definition (.api file)
type (
    CreateRequest {      // ✅ No "User" suffix when group is set
        Name     string `json:"name" validate:"required,min=2,max=50"`
        Email    string `json:"email" validate:"required,email"`
        Age      int    `json:"age" validate:"required,gte=18,lte=120"`
        Password string `json:"password" validate:"required,min=8"`
    }

    CreateResponse {
        Id      int64  `json:"id"`
        Message string `json:"message"`
    }

    GetRequest {          // ✅ No "User" suffix when group is set
        Id int64 `path:"id" validate:"required,gt=0"`
    }

    GetResponse {
        Id    int64  `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
        Age   int    `json:"age"`
    }

    ListRequest {         // ✅ No "User" suffix when group is set
        Page     int    `form:"page,default=1" validate:"gte=1"`
        PageSize int    `form:"page_size,default=10" validate:"gte=1,lte=100"`
        Keyword  string `form:"keyword,optional"`
    }

    ListResponse {
        Total int64        `json:"total"`
        List []Info        `json:"users"`
    }

    Info {
        Id    int64  `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }
)
```

**Tag Reference:**
- `json` - JSON field name
- `path` - Path parameter (e.g., `/users/:id`)
- `form` - Query parameter or form data
- `header` - HTTP header
- `validate` - Validation rules
- `optional` - Field is optional
- `default` - Default value

## Complete API Definition Example

```api
// user.api
info(
    title: "User API"
    desc: "User management API"
    author: "pzero"
    version: "v1"
    go_package: "user"  // ‼️ REQUIRED
)

type (
    CreateRequest {      // ✅ No "User" suffix needed
        Name     string `json:"name" validate:"required"`
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required,min=8"`
    }

    CreateResponse {
        Id int64 `json:"id"`
    }

    GetRequest {          // ✅ No "User" suffix needed
        Id int64 `path:"id" validate:"required,gt=0"`
    }

    GetResponse {
        Id    int64  `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }

    UpdateRequest {       // ✅ No "User" suffix needed
        Id   int64  `path:"id"`
        Name string `json:"name,optional"`
    }

    UpdateResponse {}

    DeleteRequest {       // ✅ No "User" suffix needed
        Id int64 `path:"id"`
    }

    DeleteResponse {}

    ListRequest {
        Page     int    `form:"page,default=1" validate:"gte=1"`
        PageSize int    `form:"page_size,default=10" validate:"gte=1,lte=100"`
        Keyword  string `form:"keyword,optional"`
    }

    ListResponse {
        Total int64       `json:"total"`
        Users []UserInfo  `json:"users"`
    }

    UserInfo {
        Id    int64  `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }
)

@server(
    prefix: /api/v1
    group: user                // ‼️ REQUIRED
    compact_handler: true      // ‼️ REQUIRED merge handler to one file
    middleware: Auth
)
service user-api {
    @doc "Create a new user"
    @handler Create      // ✅ No "User" suffix needed
    post /users (CreateRequest) returns (CreateResponse)

    @doc "Get user by ID"
    @handler Get         // ✅ No "User" suffix needed
    get /users/:id (GetRequest) returns (GetResponse)

    @doc "Update user"
    @handler Update      // ✅ No "User" suffix needed
    put /users/:id (UpdateRequest) returns (UpdateResponse)

    @doc "Delete user"
    @handler Delete      // ✅ No "User" suffix needed
    delete /users/:id (DeleteRequest) returns (DeleteResponse)

    @doc "List user"
    @handler List        // ✅ No "User" suffix needed
    get /users (ListRequest) returns (ListResponse)
}
```

## When to Use This Pattern

Use the standard three-layer REST pattern for:
- CRUD APIs
- RESTful web services
- API gateways
- Backend-for-frontend (BFF) services
- Microservice APIs
