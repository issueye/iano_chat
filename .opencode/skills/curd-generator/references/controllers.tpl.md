### Controller Template

```go
package controllers

import (
	"iano_chat/models"
	"iano_chat/pkg/web"
	"iano_chat/services"
	"net/http"
	"strconv"
)

type {ModelName}Controller struct {
	{modelVar}Service *services.{ModelName}Service
}

func New{ModelName}Controller({modelVar}Service *services.{ModelName}Service) *{ModelName}Controller {
	return &{ModelName}Controller{{modelVar}Service: {modelVar}Service}
}

type Create{ModelName}Request struct {
	// Add fields based on model
}

type Update{ModelName}Request struct {
	// Use pointer types with omitempty
}

func (c *{ModelName}Controller) Create(ctx *web.Context) {
	var req Create{ModelName}Request
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	{modelVar} := &models.{ModelName}{
		// Map request fields to model
	}

	if err := c.{modelVar}Service.Create({modelVar}); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, models.Success({modelVar}))
}

func (c *{ModelName}Controller) GetByID(ctx *web.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail("invalid id"))
		return
	}

	{modelVar}, err := c.{modelVar}Service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("{ModelName} not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success({modelVar}))
}

func (c *{ModelName}Controller) GetAll(ctx *web.Context) {
	{modelVar}s, err := c.{modelVar}Service.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success({modelVar}s))
}

func (c *{ModelName}Controller) Update(ctx *web.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail("invalid id"))
		return
	}

	var req Update{ModelName}Request
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	updates := make(map[string]interface{})
	// Add conditional field updates

	{modelVar}, err := c.{modelVar}Service.Update(id, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success({modelVar}))
}

func (c *{ModelName}Controller) Delete(ctx *web.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail("invalid id"))
		return
	}

	if err := c.{modelVar}Service.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "{ModelName} deleted successfully"}))
}
```