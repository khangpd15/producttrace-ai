package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/service"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type UserHandler struct {
	userService service.UserServiceInterface
}

func NewUserHandler(userService service.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	res, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(201, response.ResponseSuccess("User created successfully", res))
}

func (h *UserHandler) GetUserDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		apperror.HandleError(c, apperror.NewBadRequest("user ID is required"))
		return
	}

	res, err := h.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("User details retrieved successfully", res))
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		apperror.HandleError(c, apperror.NewBadRequest("user ID is required"))
		return
	}

	var req request.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	res, err := h.userService.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("User updated successfully", res))
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		apperror.HandleError(c, apperror.NewBadRequest("user ID is required"))
		return
	}

	err := h.userService.DeleteUser(c.Request.Context(), id)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("User deleted successfully", nil))
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	role := c.Query("role")
	status := c.Query("status")
	search := c.Query("search")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	res, err := h.userService.ListUsers(c.Request.Context(), page, limit, role, status, search)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Users list retrieved successfully", res))
}

func (h *UserHandler) SearchUsers(c *gin.Context) {
	var req request.SearchUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	res, err := h.userService.SearchUsers(c.Request.Context(), &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Search results retrieved successfully", res))
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	actorID := c.GetHeader("X-User-Id")
	if actorID == "" {
		apperror.HandleError(c, apperror.NewUnauthorized("Login required"))
		return
	}

	targetUserID := c.Query("user_id")
	if targetUserID == "" {
		targetUserID = actorID
	}

	res, err := h.userService.GetProfile(c.Request.Context(), targetUserID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Profile loaded successfully", res))
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	actorID := c.GetHeader("X-User-Id")
	if actorID == "" {
		apperror.HandleError(c, apperror.NewUnauthorized("Login required"))
		return
	}

	targetUserID := c.Param("id")
	if targetUserID == "" {
		targetUserID = actorID
	}

	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	res, err := h.userService.UpdateProfile(c.Request.Context(), actorID, targetUserID, &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Profile updated successfully", res))
}
