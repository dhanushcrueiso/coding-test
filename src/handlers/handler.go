package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dhanushcrueiso/coding-test/internal/db"

	"github.com/gofiber/fiber/v2"
)

type Handler struct{}

func (h *Handler) GetHealth(c *fiber.Ctx) error {

	return c.Status(200).JSON(fiber.Map{
		"status": "ok"})

}

func (h *Handler) SetStringData(c *fiber.Ctx) error {
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
	fmt.Println(ttl)
	err := db.DataSvc.Set(c.Params("key"), data.Value, &ttl)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to set data"})
	}
	return c.Status(200).JSON(fiber.Map{
		"message": "data set successfully"})
}

func (h *Handler) GetStringData(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "key is required"})
	}
	value, dataType, found := db.DataSvc.Get(key)
	if !found {
		return c.Status(404).JSON(fiber.Map{
			"error": "data not found"})
	}
	return c.JSON(fiber.Map{
		"key":   key,
		"value": value,
		"type":  dataType,
	})
}

func (h *Handler) UpdateStringData(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "key is required"})
	}

	var data struct {
		Value string `json:"value"`
	}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body"})
	}

	if db.DataSvc.Update(key, data.Value) {
		return c.Status(200).JSON(fiber.Map{
			"message": "data updated successfully"})
	} else {
		return c.Status(404).JSON(fiber.Map{
			"error": "data not found"})
	}

}

func (h *Handler) DeleteStringData(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "key is required"})
	}
	if db.DataSvc.Remove(key) {
		return c.Status(200).JSON(fiber.Map{
			"message": "data deleted successfully"})
	} else {
		return c.Status(404).JSON(fiber.Map{
			"error": "data not found"})
	}
}
