package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetListData(c *fiber.Ctx) error {
	key := c.Params("key")
	data, err := h.store.GetList(key)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Data Not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "List data retrieved successfully",
		"data":    data,
	})
}

func (h *Handler) SetListData(c *fiber.Ctx) error {
	key := c.Params("key")
	var ttl time.Duration = 0
	ttlStr := c.Query("ttl")
	if ttlStr != "" {
		ttlSeconds, err := strconv.Atoi(ttlStr)
		if err != nil || ttlSeconds < 0 {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid ttl value"})
		}
		ttl = time.Duration(ttlSeconds) * time.Second
	}

	if h.store.CreateList(key, ttl) {
		return c.Status(200).JSON(fiber.Map{
			"message": "List created successfully"})
	} else {
		return c.Status(400).JSON(fiber.Map{
			"error": "Failed to create list"})
	}
}

func (h *Handler) DeleteListData(c *fiber.Ctx) error {
	key := c.Params("key")
	if h.store.Remove(key) {
		return c.Status(200).JSON(fiber.Map{
			"message": "List deleted successfully"})
	} else {
		return c.Status(400).JSON(fiber.Map{"message": "Failed to delete list,key not found"})
	}

}

func (h *Handler) UpdateListData(c *fiber.Ctx) error {
	key := c.Params("key")
	operation := c.Params("operation")
	if operation == "push" {
		if c.Body() == nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "request body is required"})
		}
		var data struct {
			Value string `json:"value"`
		}
		if err := c.BodyParser(&data); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid request body"})
		}
		if h.store.Push(key, data.Value) {
			return c.Status(200).JSON(fiber.Map{
				"message": "added to list successfully"})
		} else {
			return c.Status(400).JSON(fiber.Map{
				"error": "Failed to add to list"})
		}

	} else {
		value, success := h.store.Pop(key)
		if !success {
			return c.Status(400).JSON(fiber.Map{
				"error": "Failed to pop from list or key not found"})
		}
		return c.Status(200).JSON(fiber.Map{
			"message": "Popped from list successfully",
			"data":    value})
	}
}
