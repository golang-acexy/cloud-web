# cloud-web

`cloud-web` defines the standard web-layer contracts for the golang-acexy cloud ecosystem. It builds on `starter-gin` and provides generated business services and routers with consistent CRUD routes, DTO field restrictions, pagination, response semantics, ID conversion, and optional row-level authority control.

The module does not implement persistence. A generated business service adapts `cloud-database` or another persistence implementation to the `BaseBizService` contract.

## Ecosystem Role

This module standardizes the application-facing HTTP layers above `starter-gin`. It defines what a generated or custom BizService must provide and turns that contract into reusable REST routing, validation, DTO-field control, pagination, and authority enforcement.

## Requirements

Current module Go version: `1.26.7`.

## Installation

```bash
go get github.com/golang-acexy/cloud-web
```

Applications also start the HTTP server through `starter-gin` and `starter-parent`.

## Architecture

The standard generated flow is:

```text
HTTP request
  -> starter-gin Router
  -> cloud-web BaseRouter
  -> generated BaseBizService
  -> Repository / persistence implementation
```

`BaseRouter` owns HTTP concerns:

- JSON body and path ID parsing
- request-field allowlists derived from DTO definitions
- camelCase request fields to snake_case database columns
- pagination validation
- standard REST responses
- optional authority field injection

`BaseBizService` owns business and persistence adaptation:

- converting save, modify, query, and response DTOs
- applying default sorting and non-page query limits
- calling repositories
- preserving persistence errors

## DTO Contracts

A generated resource normally defines four DTO roles:

```go
type UserSaveDTO struct {
	UserID uint64 `json:"userId"`
	Name   string `json:"name"`
}

type UserModifyDTO struct {
	UserID uint64 `json:"userId"`
	Name   string `json:"name"`
}

type UserQueryDTO struct {
	UserID uint64 `json:"userId"`
	Name   string `json:"name"`
}

type UserDTO struct {
	ID        uint64 `json:"id"`
	UserID    uint64 `json:"userId"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"createdAt"`
}
```

The generic parameters `S`, `M`, and `Q` passed to `BaseRouter` must be non-pointer struct types.

DTOs are part of the protocol contract. Field names must follow the standard mapping:

```text
Go field:       UserID
JSON field:     userId
database field: user_id
```

Custom JSON aliases are not interpreted when building field allowlists.

Database-generated fields such as IDs, creation timestamps, update timestamps, and deletion timestamps must not be declared in save DTOs. They may be present in query DTOs and response DTOs. The base modify route rejects common generated columns including `id`, `created_at`, `updated_at`, and `deleted_at`.

## Business Service Contract

A generated service implements `BaseBizService`:

```go
type BaseBizService[ID IDType, S, M, Q, D any] interface {
	MaxQuerySize() int
	DefaultOrderBy() string
	DefaultTimeRangeField() string
	AllowedTimeRangeFields() []string

	Save(save *S) (ID, error)
	SaveWithoutZeroFields(save *S) (ID, error)
	SaveBatch(saves []*S) ([]ID, error)

	BaseQueryByID(condition map[string]any) (*D, error)
	BaseQueryOne(condition map[string]any) (*D, error)
	BaseQuery(condition map[string]any) ([]*D, error)
	BaseQueryPage(query PagerDTO[map[string]any]) (Pager[D], error)
	BaseModifyByID(update, condition map[string]any) (int64, error)
	BaseRemoveByID(condition map[string]any) (int64, error)

	QueryByID(id ID) (*D, error)
	QueryByIDs(ids []ID) ([]*D, error)
	ExistsByID(id ID) (bool, error)
	QueryOneByCond(condition Q) (*D, error)
	QueryByCond(condition Q) ([]*D, error)
	CountByCond(condition Q) (int64, error)
	CountByMap(condition map[string]any) (int64, error)
	QueryPage(pager PagerDTO[Q]) (Pager[D], error)

	ModifyByID(id ID, updated *M) (int64, error)
	ModifyByIDWithoutZeroFields(id ID, updated *M) (int64, error)
	ModifyByIDWithMap(id ID, updated map[string]any) (int64, error)
	ModifyByCond(condition Q, updated *M) (int64, error)
	ModifyByMap(updated, condition map[string]any) (int64, error)

	RemoveByID(id ID) (int64, error)
	RemoveByIDs(ids []ID) (int64, error)
	RemoveByCond(condition Q) (int64, error)
	RemoveByMap(condition map[string]any) (int64, error)
}
```

The `Base*` methods receive database-column maps from `BaseRouter`. The other methods form the typed business API available to application code. Map-based operations preserve explicit zero values and are the preferred bridge for generic REST conditions and updates.

All persistence failures must be returned as errors. Generated implementations must not collapse database errors into `false`, nil results, or zero values.

## Base Router

Create a router by embedding `BaseRouter` in a resource-specific Gin router:

```go
type UserRouter struct {
	*webcloud.BaseRouter[
		uint64,
		UserSaveDTO,
		UserModifyDTO,
		UserQueryDTO,
		UserDTO,
	]
}

func NewUserRouter(
	service webcloud.BaseBizService[
		uint64,
		UserSaveDTO,
		UserModifyDTO,
		UserQueryDTO,
		UserDTO,
	],
) *UserRouter {
	return &UserRouter{
		BaseRouter: webcloud.NewBaseRouter[
			uint64,
			UserSaveDTO,
			UserModifyDTO,
			UserQueryDTO,
			UserDTO,
		](service),
	}
}

func (r *UserRouter) Info() *ginstarter.RouterInfo {
	return &ginstarter.RouterInfo{GroupPath: "/users"}
}

func (r *UserRouter) Handlers(router *ginstarter.RouterWrapper) {
	r.RegisterBaseHandlers(router, r)
	// Register resource-specific routes here.
}
```

The second argument is the concrete router used for handler dispatch. Methods
implemented by `UserRouter` override the promoted `BaseRouter` methods; methods
that are not overridden continue to use the embedded default implementations.

`RegisterBaseHandlers` adds these routes relative to `GroupPath`:

| Method | Path | Behavior |
| --- | --- | --- |
| `POST` | `/save` | Save one resource. |
| `GET` | `/by-id/:id` | Query one resource by ID. |
| `POST` | `/query-one` | Query one resource by condition. |
| `POST` | `/query` | Query a bounded list by condition. |
| `POST` | `/query-by-page` | Query one validated page. |
| `PUT` | `/by-id/:id` | Modify one resource by ID. |
| `DELETE` | `/by-id/:id` | Remove one resource by ID. |

JSON media type is required for routes with request bodies.

## Authority-Controlled Router

Use `NewBaseRouterWithAuthority` when every base data operation must be restricted to the authenticated identity:

```go
authorityFetch := func(request *ginstarter.Request) webcloud.Authority[uint64] {
	return currentAuthority(request)
}

baseRouter := webcloud.NewBaseRouterWithAuthority[
	uint64,
	UserSaveDTO,
	UserModifyDTO,
	UserQueryDTO,
	UserDTO,
](
	service,
	authorityFetch,
	webcloud.AuthorityDataField{
		StructField: "UserID",
		Column:      "user_id",
	},
)
```

`AuthorityDataField` separates the two names that cannot be safely inferred from each other:

- `StructField` is the Go field used to inject identity into the save DTO.
- `Column` is the database field added to query, update, and delete conditions.

Authority values are written after client input is parsed and converted. Client-supplied values for the controlled field are therefore overwritten by the authenticated identity.

Behavior by operation:

| Operation | Authority behavior |
| --- | --- |
| Save | Overwrites the DTO `StructField`. |
| Query | Overwrites the condition `Column`. |
| Page query | Adds the condition after parsing the client condition. |
| Modify | Restricts the target condition and overwrites the update column. |
| Remove | Restricts the delete condition. |

Missing authority configuration is treated as a system error when the protected operation is used. A configured authority fetcher returning nil is treated as an unauthorized request.

Use the explicit methods when custom routes need authority data:

```go
authority := router.GetAuthorityData(request)
optional := router.GetOptionalAuthorityData(request)
```

`GetAuthorityData` requires an authenticated identity. `GetOptionalAuthorityData` allows the configured fetcher to return nil.

## Pagination

The page request contract is:

```go
type TimeRange struct {
	Field string         `json:"field"`
	Start json.Timestamp `json:"start"`
	End   json.Timestamp `json:"end"`
}

type PagerDTO[T any] struct {
	Size       int         `json:"size" binding:"required,gte=1,lte=2000"`
	Number     int         `json:"number" binding:"required,gte=1"`
	Condition  T           `json:"condition"`
	TimeRanges []TimeRange `json:"timeRanges"`
}
```

Example request:

```json
{
  "number": 1,
  "size": 20,
  "condition": {
    "name": "Alice"
  },
  "timeRanges": [
    {
      "start": 1785513600000,
      "end": 1785600000000
    },
    {
      "field": "updatedAt",
      "start": 1785513600000
    }
  ]
}
```

An omitted `field` uses `DefaultTimeRangeField`; at most one range may omit it. Explicit fields are converted to snake case and must match `AllowedTimeRangeFields`. A range must contain at least one bound, and when both bounds are present it uses `[start, end)` and requires `start < end`.

The response contains page metadata and records:

```go
type Pager[T any] struct {
	Records []*T  `json:"records"`
	Total   int64 `json:"total"`
	Size    int   `json:"size"`
	Number  int   `json:"number"`
}
```

`number` must be at least 1. `size` must be between 1 and 2000.

## ID Types

Supported ID underlying types are:

```go
int, uint, int32, uint32, int64, uint64, string
```

Named types are supported:

```go
type UserID int64

id, err := webcloud.ConvertStringToID[UserID]("1001")
```

An invalid path ID is returned as a bad-parameters REST response instead of a system exception.

## Embedded DTO Fields

Anonymous embedded structs are expanded when query and modify allowlists are built:

```go
type CommonQuery struct {
	CreatedAt int64
}

type UserQueryDTO struct {
	CommonQuery
	UserID uint64
}
```

The resulting database fields are `created_at` and `user_id`.

Rules:

- Top-level DTOs must be non-pointer structs.
- Anonymous struct values and anonymous struct pointers are expanded.
- Shallower fields override deeper embedded fields.
- Conflicting fields or database columns at the same depth fail during router construction.
- Recursive anonymous embedding is cycle-safe.
- Unexported fields are ignored.

## Response Semantics

Base routes use the standard `starter-gin` REST envelope.

| Situation | Result |
| --- | --- |
| Invalid body, page, field, or path ID | Bad parameters. |
| Single query has no record | Success with empty data. |
| List query has no records | Success with `[]`, as returned by the service. |
| Modify or remove affects zero rows | Bad parameters. |
| Business or persistence method returns an error | System exception pipeline. |
| Required authority is absent | Unauthorized. |

The service controls the returned list value. Generated services should return an empty non-nil slice when no list records exist.

## Starter Integration

Register generated routers with `GinStarter` and manage lifecycle through the parent loader:

```go
starter := &ginstarter.GinStarter{
	Config: ginstarter.GinConfig{
		ListenAddress: ":8080",
		Routers: []ginstarter.Router{
			NewUserRouter(userService),
		},
	},
}

loader := parent.InitStarterLoader([]parent.Starter{starter})
if err := loader.Start(); err != nil {
	panic(err)
}
```

Shut down through the same loader so Gin participates in the shared application lifecycle:

```go
_, err := loader.StopAllBySetting(10 * time.Second)
if err != nil {
	panic(err)
}
```

## Important Errors

| Error | Meaning |
| --- | --- |
| `ErrUnauthorized` | A required authority identity is absent. |
| `ErrBadRequestParameters` | Parsed request fields are empty or invalid. |
| `ErrUnsupportedIDType` | The ID type is outside `IDType`. |
| `ErrAuthorityFetchRequired` | A protected route has no authority fetcher. |
| `ErrAuthorityStructFieldRequired` | Save authority injection has no Go field name. |
| `ErrAuthorityDataColumnRequired` | Protected database operations have no authority column. |
| `ErrDTOTypeMustBeStruct` | A query or modify DTO is not a struct. |
| `ErrEmbeddedFieldConflict` | Embedded DTO fields resolve ambiguously. |

## Testing

Run tests from the module directory:

```bash
GOMODCACHE=/Users/acexy/Repository/cache/golang go test ./...
```

The test suite uses `httptest` against the actual Gin engine and does not require a fixed listening port or external infrastructure.
