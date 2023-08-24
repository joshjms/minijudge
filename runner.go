package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func RunTask(port int, req *RunRequest) (RunResponse, error) {
	resp, err := StartContainer(port)
	if err != nil {
		return RunResponse{}, err
	}

	type ExecuteResult struct {
		runResp RunResponse
		err     error
	}

	result := make(chan ExecuteResult, 1)

	go func() {
		fmt.Println("Running code")
		runResp, err := RunCode(port, req)
		result <- ExecuteResult{runResp, err}
	}()

	defer StopContainer(resp, port)

	select {
	case <-time.After(5 * time.Second):
		return RunResponse{}, fmt.Errorf("execution timed out")
	case res := <-result:
		return res.runResp, res.err
	}
}

func RunCode(port int, runReq *RunRequest) (RunResponse, error) {
	runResp := RunResponse{
		Message: "Error",
	}

	payloadBytes, err := json.Marshal(runReq)
	if err != nil {
		return runResp, err
	}

	req, err := http.NewRequest("GET", "http://localhost:"+strconv.Itoa(port), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return runResp, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return runResp, err
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		return runResp, err
	}

	return runResp, nil
}
