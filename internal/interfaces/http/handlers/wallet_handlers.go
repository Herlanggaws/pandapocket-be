package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"panda-pocket/internal/application/finance"

	"github.com/gin-gonic/gin"
)

type WalletHandlers struct {
	createWalletUseCase *finance.CreateWalletUseCase
	updateWalletUseCase *finance.UpdateWalletUseCase
	getWalletsUseCase   *finance.GetWalletsUseCase
	deleteWalletUseCase *finance.DeleteWalletUseCase
}

func NewWalletHandlers(
	createWalletUseCase *finance.CreateWalletUseCase,
	updateWalletUseCase *finance.UpdateWalletUseCase,
	getWalletsUseCase *finance.GetWalletsUseCase,
	deleteWalletUseCase *finance.DeleteWalletUseCase,
) *WalletHandlers {
	return &WalletHandlers{
		createWalletUseCase: createWalletUseCase,
		updateWalletUseCase: updateWalletUseCase,
		getWalletsUseCase:   getWalletsUseCase,
		deleteWalletUseCase: deleteWalletUseCase,
	}
}

// CreateWallet handles wallet creation
func (h *WalletHandlers) CreateWallet(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req finance.CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := h.createWalletUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusCreated, gin.H{
		"wallet": response,
	})
}

// UpdateWallet handles wallet updates
func (h *WalletHandlers) UpdateWallet(c *gin.Context) {
	userID := c.GetInt("user_id")
	walletIDStr := c.Param("id")

	var walletID int
	if _, err := fmt.Sscanf(walletIDStr, "%d", &walletID); err != nil {
		BadRequestResponse(c, "INVALID_WALLET_ID", "Invalid wallet ID")
		return
	}

	var req finance.UpdateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := h.updateWalletUseCase.Execute(c.Request.Context(), userID, walletID, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"wallet": response,
	})
}

// GetWallets handles getting wallets
func (h *WalletHandlers) GetWallets(c *gin.Context) {
	userID := c.GetInt("user_id")

	req := finance.GetWalletsRequest{
		Search: c.Query("search"),
	}

	if pageParam := c.Query("page"); pageParam != "" {
		if page, err := strconv.Atoi(pageParam); err == nil {
			req.Page = page
		}
	}

	if limitParam := c.Query("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil {
			req.Limit = limit
		}
	}

	response, err := h.getWalletsUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		InternalServerErrorResponse(c, "FETCH_WALLETS_ERROR", "Failed to fetch wallets")
		return
	}

	SuccessResponse(c, http.StatusOK, response)
}

// DeleteWallet handles wallet deletion
func (h *WalletHandlers) DeleteWallet(c *gin.Context) {
	userID := c.GetInt("user_id")
	walletIDStr := c.Param("id")

	var walletID int
	if _, err := fmt.Sscanf(walletIDStr, "%d", &walletID); err != nil {
		BadRequestResponse(c, "INVALID_WALLET_ID", "Invalid wallet ID")
		return
	}

	err := h.deleteWalletUseCase.Execute(c.Request.Context(), userID, walletID)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Wallet deleted successfully",
	})
}
