package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"idm/inner/web"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Test that registerMiddleware applies the recover middleware correctly
func TestRegisterMiddleware_Recovery(t *testing.T) {
	server := web.NewServer()
	RegisterMiddleware(server.App)

	// Register a route that panics
	server.App.Get("/panic", func(c *fiber.Ctx) error {
		panic("intentional panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := server.App.Test(req)
	assert.NoError(t, err, "app.Test should not return an error")
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "Panic should be recovered to 500")
}

// Test that normal routes work alongside middleware
func TestRegisterMiddleware_NormalFlow(t *testing.T) {
	server := web.NewServer()
	RegisterMiddleware(server.App)

	// Register a normal route
	server.App.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := server.App.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "ok", string(body))
}

// Test that RequestID middleware sets X-Request-ID header
func TestRegisterMiddleware_RequestIDHeader(t *testing.T) {
	server := web.NewServer()
	RegisterMiddleware(server.App)

	// Register a dummy route
	server.App.Get("/testid", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/testid", nil)
	resp, err := server.App.Test(req)
	assert.NoError(t, err, "app.Test should not return an error")
	// Verify header is present
	xid := resp.Header.Get("X-Request-ID")
	assert.NotEmpty(t, xid, "Expected X-Request-ID header to be set")
}
