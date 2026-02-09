---
name: "curd-generator"
description: "Generates complete CRUD code for Go web applications based on GORM models. Invoke when user asks to generate CRUD operations for a model or create controller/service/route code."
---

# CURD Generator Skill

This skill generates complete CRUD (Create, Read, Update, Delete) code for a Go web application based on model definitions.

## Project Structure

The generated code follows the project structure:
- `models/` - Data models with GORM tags
- `controllers/` - HTTP handlers with request/response structs
- `services/` - Business logic layer
- `routes/` - Route registration

## Code Conventions

### Model Conventions
- Embed `BaseModel` for ID (string UUID), CreatedAt, UpdatedAt
- Use `gorm` tags for column definitions
- Implement `TableName()` method
- Use `int64` for primary key when model uses auto-increment ID
- Package: `models`

### Controller Conventions
- Request structs with JSON binding tags
- Update requests use pointer types with `omitempty`
- Use `ctx.BindAndValidate()` for creation
- Use `ctx.Bind()` for updates
- Return `models.Success()` or `models.Fail()`
- Use HTTP status codes appropriately (201 for create, 200 for others)
- ID parameters parsed with `strconv.ParseInt()` for int64 IDs
- Package: `controllers`

### Service Conventions
- Inject `*gorm.DB` via constructor
- Return errors directly (no wrapping for simple cases)
- Use `s.db.Create()`, `s.db.First()`, `s.db.Find()`, `s.db.Updates()`, `s.db.Delete()`
- Support `GetByUserID()`, `GetByStatus()` query methods
- Support `DeleteByUserID()` for batch delete
- Support `Count()`, `CountByUserID()` statistics methods
- Package: `services`

### Route Conventions
- Setup in `routes/routes.go`
- Pattern: `/api/{pluralResource}/:id` for single resource operations
- Pattern: `/api/{pluralResource}` for list operations
- Query-based filters: `/api/{pluralResource}/user?user_id=xxx`
- Special endpoints: `/api/{pluralResource}/:id/config` for config operations
- Middleware: Recovery, Logger, CORS already configured

## Template Code

`./references/service.tpl.md`
`./references/controller.tpl.md`
`./references/router.tpl.md`

## Input Format

Provide model definition with:
1. Model name (e.g., "Session")
2. Fields with:
   - Name (Go field name)
   - Type (Go type)
   - JSON tag
   - GORM tag (optional)
   - Description (optional)

## Output

Generate:
1. **Service code** - Business logic methods
2. **Controller code** - Request/Response structs + CRUD handler methods
3. **Route registration** - Add to SetupRoutes function

## Example Usage

Input:
```
Generate CRUD for "Session" model:
- ID: int64 (primary key, auto-increment)
- UserID: int64 (gorm:"index;not null")
- Title: string (gorm:"size:255;not null")
- Status: SessionStatus (gorm:"size:20;default:'active'")
- ConfigJSON: string (gorm:"type:text")
- MessageCount: int (gorm:"default:0")
- TotalTokens: int (gorm:"default:0")
- CreatedAt: time.Time
- UpdatedAt: time.Time
- LastActiveAt: time.Time
```

Generated code follows existing Agent/Message model patterns exactly.
