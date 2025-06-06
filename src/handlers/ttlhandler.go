package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetTtlData(c *fiber.Ctx) error {
	key := c.Params("key")

	ttl, bool := h.store.GetTTL(key)
	if !bool {
		return c.Status(400).JSON(fiber.Map{
			"error": "Failed to set TTL/data not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "TTL set successfully",
		"data":    int(ttl.Seconds()),
	})
}

func (h *Handler) SetTtlData(c *fiber.Ctx) error {
	key := c.Params("key")
	ttlStr := c.Query("ttl")
	if ttlStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid ttl value"})
	}

	ttlSeconds, err := strconv.Atoi(ttlStr)
	if err != nil || ttlSeconds < 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid ttl value"})
	}
	ttl := time.Duration(ttlSeconds) * time.Second

	if h.store.SetTTL(key, ttl) {
		return c.Status(200).JSON(fiber.Map{
			"message": "ttl updated"})
	} else {
		return c.Status(400).JSON(fiber.Map{
			"error": "key not found"})
	}
}
