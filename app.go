package main

import (
	"errors"
	"log"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"

	"github.com/ilievs/fibers/core"
	"github.com/ilievs/fibers/mqtt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	mochi "github.com/mochi-mqtt/server/v2"
)

var credentials = map[string]string{
	"chicho": "petyo1",
}

var tokens = make(map[string]interface{})

type UserPass struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func handleAuth(c echo.Context) error {
	userPass := new(UserPass)
	c.Bind(userPass)

	pass, ok := credentials[userPass.Username]
	if !ok {
		log.Println("missing in credentials")
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	if userPass.Password != pass {
		log.Println("pass doesn't match")
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	authToken := strconv.Itoa(int(rand.Int64())) + strconv.Itoa(int(rand.Int64()))
	tokens[authToken] = 1
	return c.JSON(http.StatusOK, map[string]string{
		"access_token": authToken,
	})
}

func validateAuth(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return errors.New("no Authorization header in request")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return errors.New("invalid Authorization header")
	}

	token := parts[1]
	_, ok := tokens[token]
	if !ok {
		return errors.New("invalid token")
	}

	return nil
}

var deviceStateCache = make(map[string]*core.State)

func getDeviceState(c echo.Context) error {
	err := validateAuth(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}
	deviceId := c.Param("deviceId")
	return c.JSON(http.StatusOK, deviceStateCache[deviceId])
}

var deviceMan = core.NewBasicDeviceManager()

func sendDeviceCommand(c echo.Context) error {
	err := validateAuth(c)
	if err != nil {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}
	deviceId := c.Param("deviceId")
	command := new(core.Command)
	c.Bind(command)
	err = deviceMan.SendCommand(deviceId, command)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func RunApplication() {
	server := mochi.New(&mochi.Options{
		InlineClient: true,
	})

	mqttClient := mqtt.NewMochiClient(server)

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
	e.POST("/auth/token", handleAuth)
	e.GET("/devices/:deviceId/stats", getDeviceState)
	e.POST("/devices/:deviceId/command", sendDeviceCommand)

	e.Static("/", "static")

	// Start server
	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}

}
