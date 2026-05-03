# Error Handling Architecture

## Overview

OMNI implements a comprehensive error handling flow that propagates errors from the database layer through services and handlers to HTTP responses, and finally to the frontend for user-facing display.

**Error Flow:**
\\\
Database Error ? Repository ? Service ? Handler ? HTTP Response ? Frontend ? User Notification
\\\

---

## Backend Error Handling

### 1. Error Types & Constants

**Location:** \ackend/internal/errors/errors.go\

The backend defines a standardized \AppError\ type with predefined error categories:

\\\go
type AppError struct {
    Code        int         // HTTP status code
    Type        string      // Error category (e.g., \ VALIDATION_ERROR\)
    Message     string      // User-facing message
    Details     interface{} // Additional context (omitted in production)
    InternalErr error       // Wrapped internal error (not exposed to client)
}
\\\

**Error Type Constants:**
- \TypeValidation\ ? HTTP 400 (Bad Request)
- \TypeNotFound\ ? HTTP 404 (Not Found)
- \TypeUnauthorized\ ? HTTP 401 (Unauthorized)
- \TypeForbidden\ ? HTTP 403 (Forbidden)
- \TypeConflict\ ? HTTP 409 (Conflict)
- \TypeRateLimit\ ? HTTP 429 (Too Many Requests)
- \TypeInternal\ ? HTTP 500 (Internal Server Error)
- \TypeBadRequest\ ? HTTP 400 (Bad Request)
- \TypeServiceUnavail\ ? HTTP 503 (Service Unavailable)

**Constructor Functions:**
\\\go
// Validation errors
errors.NewValidationError(message, details)

// Resource not found
errors.NewNotFoundError(resource)

// Authentication/Authorization
errors.NewUnauthorizedError(message)
errors.NewForbiddenError(message)

// Conflict (e.g., duplicate resource)
errors.NewConflictError(message)

// Rate limiting
errors.NewRateLimitError()

// Internal errors
errors.NewInternalError(err)

// Bad request
errors.NewBadRequestError(message)

// Service unavailable
errors.NewServiceUnavailableError(service)

// Wrap existing error with context
errors.Wrap(err, \additional context\)
\\\

### 2. Error Middleware

**Location:** \ackend/internal/middleware/error_handler.go\

The \ErrorHandler\ middleware intercepts all errors and panics:

\\\go
type ErrorResponse struct {
    Success bool        \json:\success\\
    Error   string      \json:\error\\
    Type    string      \json:\type omitempty\\
    Details interface{} \json:\details omitempty\\
}
\\\

**Behavior:**
- **Panics:** Caught and logged with stack trace (dev only)
- **AppErrors:** Converted to HTTP response with appropriate status code
- **Unhandled Errors:** Logged as warnings, returned as 500 Internal Server Error
- **Production Mode:** Details field omitted to prevent information leakage

**Example Response:**
\\\json
{
  \success\: false,
  \error\: \Order not found\,
  \type\: \NOT_FOUND\
}
\\\

### 3. Standard API Response Format

**Location:** \ackend/internal/dto/response/api_response.go\

All API responses follow a consistent structure:

\\\go
type APIResponse struct {
    Success bool        \json:\success\\
    Data    interface{} \json:\data omitempty\\
    Error   string      \json:\error omitempty\\
    Message string      \json:\message omitempty\\
    Meta    *Meta       \json:\meta omitempty\\
}

type Meta struct {
    Total      int \json:\total omitempty\\
    Page       int \json:\page omitempty\\
    PageSize   int \json:\page_size omitempty\\
    TotalPages int \json:\total_pages omitempty\\
}
\\\

**Response Builders:**
\\\go
// Success response
response.Success(data)

// Success with pagination
response.SuccessWithMeta(data, &Meta{Total: 100, Page: 1})

// Error response
response.Error(\User not found\)

// Error with additional message
response.ErrorWithMessage(err, \Failed to process order\)

// Error with detail
response.ErrorWithDetail(summary, detail)

// Partial success (HTTP 207 - mixed results)
response.PartialSuccess(data)

// Error with data (e.g., per-item results)
response.ErrorWithData(msg, data)
\\\

### 4. Handler Response Helpers

**Location:** \ackend/internal/handlers/response_helpers.go\

Handlers use convenience functions to send responses:

\\\go
// Success responses
respondSuccess(c, data, count)        // HTTP 200 with data
respondCreated(c, data)               // HTTP 201 with data
respondWithConfig(c, config)          // HTTP 200 with single config
respondWithConfigs(c, configs)        // HTTP 200 with multiple configs
respondDeleted(c, msg)                // HTTP 200 with message

// Error responses
respondBadRequest(c, msg)             // HTTP 400
respondUnauthorized(c, msg)           // HTTP 401
respondNotFound(c, msg)               // HTTP 404
respondInternalError(c, err)          // HTTP 500
respondServiceUnavailable(c, msg)     // HTTP 503
\\\

### 5. Handler Pattern

**Location:** \ackend/internal/handlers/\ (141+ files)

Handlers follow a strict pattern:

\\\go
func (h *OrderHandler) GetOrder(c *gin.Context) {
    // 1. Extract tenant from middleware context
    tenantID := c.GetString(\tenant_id\)
    if tenantID == \\ {
        respondBadRequest(c, \Missing tenant_id\)
        return
    }

    // 2. Parse and validate request
    var req GetOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondBadRequest(c, err.Error())
        return
    }

    // 3. Call service layer (NO business logic here)
    order, err := h.orderService.GetOrder(c.Request.Context(), tenantID, req.OrderID)
    if err != nil {
        // Service returns AppError or wrapped error
        if appErr, ok := err.(*errors.AppError); ok {
            c.Error(appErr)  // Middleware handles it
        } else {
            respondInternalError(c, err)
        }
        return
    }

    // 4. Return success response
    respondSuccess(c, order, 1)
}
\\\

**Key Rules:**
- ? Extract tenant from middleware context
- ? Validate request input
- ? Call service layer only
- ? Let middleware handle AppErrors
- ? Return consistent response format
- ? NO database access
- ? NO business logic
- ? NO default tenant

### 6. Service Layer Error Handling

Services wrap errors with context and return AppErrors:

\\\go
func (s *OrderService) GetOrder(ctx context.Context, tenantID, orderID string) (*Order, error) {
    // Database error
    order, err := s.repo.GetOrder(ctx, tenantID, orderID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.NewNotFoundError(\Order\)
        }
        return nil, errors.Wrap(err, \failed to fetch order\)
    }

    // Validation error
    if order.Status == \\ {
        return nil, errors.NewValidationError(\Invalid order status\, nil)
    }

    return order, nil
}
\\\

### 7. Platform-Specific Errors

**Location:** \ackend/internal/dto/response/platform_error.go\

Third-party platform errors are wrapped:

\\\go
type PlatformError struct {
    Platform string \json:\platform\\
    Code     string \json:\code\\
    Message  string \json:\message\\
}

// Usage
response.ErrorWithPlatform(\shopee\, \INVALID_PRODUCT\, \Product not found on Shopee\)
\\\

---

## Frontend Error Handling

### 1. API Client Setup

**Location:** \rontend/src/api/client.ts\

The Axios client includes request and response interceptors:

\\\	ypescript
export interface ApiResponse<T = unknown> {
    success: boolean;
    data?: T;
    error?: string;
    message?: string;
}
\\\

**Request Interceptor:**
- Adds JWT token from auth store
- Adds tenant ID header (\x-tenant-id\)
- Adds CSRF token for mutations
- Logs requests in dev mode

**Response Interceptor:**
- Delegates to \handleResponseError\ for errors
- Logs responses in dev mode

### 2. Error Handler

**Location:** \rontend/src/api/clientErrorHandler.ts\

Centralized error handling for all HTTP responses:

\\\	ypescript
export function handleResponseError(error: AxiosError): Promise<never>
\\\

**Handles:**

| Scenario | Action |
|----------|--------|
| **Timeout** | Reject with \Request timeout\ message |
| **Network Error** | Reject with \Cannot connect to server\ message |
| **401 Unauthorized** | Clear auth, redirect to login |
| **403 Forbidden** | Show \Permission denied\ toast, reject |
| **500 Server Error** | Show \Server error\ toast, reject |
| **Other Errors** | Sanitize message, reject |

**Key Function:**
\\\	ypescript
export function handleAuthExpired(): void {
    useAuthStore.getState().clearAuth();
    window.location.href = \/login?returnUrl=\\;
}
\\\

### 3. Error Boundary Component

**Location:** \rontend/src/components/ErrorBoundary.tsx\

React Error Boundary catches component rendering errors:

\\\	ypescript
export class ErrorBoundary extends Component<Props, State> {
    static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error };
    }

    componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        logger.error(\ErrorBoundary caught an error\, {
            error: error.message,
            stack: error.stack,
            componentStack: errorInfo.componentStack,
        });
    }

    render() {
        if (this.state.hasError) {
            return (
                <Result
                    status=\error\
                    title=\Something went wrong\
                    subTitle={isDev ? error.message : \Please reload the page\}
                    extra={<Button onClick={this.handleReset}>Reload Page</Button>}
                />
            );
        }
        return this.props.children;
    }
}
\\\

### 4. Notification Store

**Location:** \rontend/src/stores/notificationStore.ts\

Centralized notification management:

\\\	ypescript
export type NotificationType = \success\ | \error\ | \warning\ | \info\;
export type NotificationCategory = \sync\ | \order\ | \product\ | \inventory\ | \auth\ | \system\ | \export\;

export interface NotificationItem {
    id: string;
    type: NotificationType;
    category: NotificationCategory;
    title: string;
    message: string;
    timestamp: number;
    read: boolean;
    actionUrl?: string;
}
\\\

**Actions:**
\\\	ypescript
addNotification(item)      // Add notification
markAsRead(id)             // Mark as read
markAllAsRead()            // Mark all as read
removeNotification(id)     // Remove notification
clearAll()                 // Clear all notifications
toggleDropdown()           // Toggle notification dropdown
closeDropdown()            // Close dropdown
\\\

**Usage:**
\\\	ypescript
useNotificationStore.getState().addNotification({
    type: \error\,
    category: \order\,
    title: \Order Sync Failed\,
    message: \Failed to sync orders from Shopee\,
});
\\\

### 5. Ant Design Message API

**Location:** \rontend/src/components/AntStaticApi.ts\

Static Ant Design message instance for toast notifications:

\\\	ypescript
import { message } from \@/components/AntStaticApi\;

// Show error toast
message.error(\Failed to save order\);

// Show success toast
message.success(\Order saved successfully\);

// Show warning toast
message.warning(\This action cannot be undone\);

// Show info toast
message.info(\Processing your request...\);
\\\

---

## Summary

OMNI'\''s error handling ensures:
- ? Consistent error format across all endpoints
- ? Appropriate HTTP status codes
- ? User-friendly error messages
- ? Secure error details (hidden in production)
- ? Proper error propagation through layers
- ? Frontend error interception and display
- ? Auth expiration handling
- ? Network error resilience
