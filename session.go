package main

func getUserFromSession(sessionId string) (string, error) {
	userData, err := redisClient.HGet(sessionId, "user_data").Result()
	if userData != "" {
		return userData, nil
	}
	return "", err
}
