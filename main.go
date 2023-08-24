package main

import (
	"fmt"
	"math"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

type (
	RunRequest struct {
		Code string `json:"code"`
	}

	RunResponse struct {
		Message      string `json:"message"`
		Error        string `json:"error"`
		Stdout       string `json:"stdout"`
		Stderr       string `json:"stderr"`
		ExecDuration int64  `json:"exec_duration"`
		MemUsage     int64  `json:"mem_usage"`
	}

	PortManager struct {
		mutex     sync.Mutex
		nextPort  int
		maxPort   int
		usedPorts map[int]bool
	}
)

func InitializePortManager(startPort, maxPort int) *PortManager {
	return &PortManager{
		nextPort:  startPort,
		maxPort:   maxPort,
		usedPorts: make(map[int]bool),
	}
}

func (pm *PortManager) GetAvailablePort() (int, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for port := pm.nextPort; port <= pm.maxPort; port++ {
		if !pm.usedPorts[port] {
			pm.usedPorts[port] = true
			pm.nextPort = port + 1
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports")
}

func (pm *PortManager) ReleasePort(port int) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	delete(pm.usedPorts, port)
	pm.nextPort = int(math.Min(float64(pm.nextPort), float64(port)))
}

var portManager *PortManager

func main() {
	InitializeDocker()
	portManager = InitializePortManager(30000, 30100)

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		port, err := portManager.GetAvailablePort()
		if err != nil {
			return c.String(http.StatusInternalServerError, "No available ports")
		}

		fmt.Println("Running container on port", port)

		req := new(RunRequest)
		if err := c.Bind(req); err != nil {
			return c.String(http.StatusBadRequest, "Invalid request")
		}

		resp, err := RunTask(port, req)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(200, resp)
	})

	e.Logger.Fatal(e.Start(":3000"))
}
