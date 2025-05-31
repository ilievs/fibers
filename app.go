package main

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/ilievs/fibers/core"
	"github.com/ilievs/fibers/mqtt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	mochi "github.com/mochi-mqtt/server/v2"
)

var credentials = map[string]interface{}{
	"chicho:petyo": 1,
}

var sessions = map[string]interface{}{}

// Handler
func handleLogin(c echo.Context) error {

	return c.String(http.StatusOK, "Hello, World!")
}

func RunApplication() {
	server := mochi.New(&mochi.Options{
		InlineClient: true,
	})

	deviceMan := core.NewBasicDeviceManager()
	mqttClient := mqtt.NewMochiClient(server)
	
	deviceStateCache := make(map[string]*core.State)

	stateChangesChan := deviceMan.SubscribeToStateChanges()
	go func() {
		for dev := range stateChangesChan {
			deviceStateCache[dev.Id()] = dev.GetState()
		}
	}()

	broker := mqtt.NewMochiBroker(server)
	broker.Start(
		[]mochi.Hook{new(mqtt.AddNewDeviceHook)},
		[]any{&mqtt.HookOptions{MqttClient: mqttClient, DeviceManager: deviceMan}})

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/login", handleLogin)
	e.GET("/devices/:deviceId/stats", func(c echo.Context) error {
		deviceId := c.Param("deviceId")
		return c.JSON(http.StatusOK, deviceStateCache[deviceId])
	})

	e.POST("/devices/:deviceId/command", func (c echo.Context) error {
		deviceId := c.Param("deviceId")
		command := new(core.Command)
		c.Bind(command)
		err := deviceMan.SendCommand(deviceId, command)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		return c.NoContent(http.StatusOK)
	})

	// Start server
	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
	
}
