package handlers

import (
	"errors"
	"io"
	"net/http"

	appMCP "panda-pocket/internal/application/mcp"
	"panda-pocket/internal/domain/entitlement"

	"github.com/gin-gonic/gin"
)

const mcpMaxBodyBytes = 1 << 20

type MCPHandlers struct {
	tokens *appMCP.TokenService
	server *appMCP.Server
}

func NewMCPHandlers(tokens *appMCP.TokenService, server *appMCP.Server) *MCPHandlers {
	return &MCPHandlers{tokens: tokens, server: server}
}

func (h *MCPHandlers) GetToken(c *gin.Context) {
	status, err := h.tokens.Status(c.Request.Context(), c.GetInt("user_id"))
	if err != nil {
		h.writeTokenErr(c, err)
		return
	}
	SuccessResponse(c, http.StatusOK, status)
}

func (h *MCPHandlers) IssueToken(c *gin.Context) {
	issued, err := h.tokens.Issue(c.Request.Context(), c.GetInt("user_id"))
	if err != nil {
		h.writeTokenErr(c, err)
		return
	}
	SuccessResponse(c, http.StatusOK, issued)
}

func (h *MCPHandlers) RevokeToken(c *gin.Context) {
	if err := h.tokens.Revoke(c.Request.Context(), c.GetInt("user_id")); err != nil {
		h.writeTokenErr(c, err)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"active": false})
}

func (h *MCPHandlers) Serve(c *gin.Context) {
	plain := bearerToken(c.GetHeader("Authorization"))
	userID, err := h.tokens.Authenticate(c.Request.Context(), plain)
	if err != nil {
		UnauthorizedResponse(c, "INVALID_TOKEN", "Invalid token")
		return
	}
	c.Request.Body = http.MaxBytesReader(nil, c.Request.Body, mcpMaxBodyBytes)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			SendErrorResponse(c, http.StatusRequestEntityTooLarge, "MCP_BODY_TOO_LARGE", "MCP request body exceeds 1 MB")
			return
		}
		BadRequestResponse(c, "MCP_BAD_REQUEST", "Could not read the MCP request")
		return
	}
	response, isNotification, err := h.server.Dispatch(c.Request.Context(), userID, body)
	if err != nil {
		InternalServerErrorResponse(c, "MCP_ERROR", "MCP request failed")
		return
	}
	if isNotification {
		c.Status(http.StatusAccepted)
		return
	}
	c.Data(http.StatusOK, "application/json", response)
}

func (h *MCPHandlers) writeTokenErr(c *gin.Context, err error) {
	if errors.Is(err, entitlement.ErrPremiumRequired) {
		PremiumRequiredResponse(c, err)
		return
	}
	InternalServerErrorResponse(c, "MCP_TOKEN_ERROR", "Could not update the MCP token")
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) > len(prefix) && header[:len(prefix)] == prefix {
		return header[len(prefix):]
	}
	return ""
}
