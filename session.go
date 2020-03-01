package main

import "errors"

func getUserFromSession(sessionId string) (string, error) {
	userData, _ := redisClient.HGet(sessionId, "user_data").Result()
	if userData != "" {
		return userData, nil
	}
	return "", errors.New("session not found")
}
