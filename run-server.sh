#!/bin/bash

export SECRET_KEY="super-secure-secret-key"
export API_KEY="super-secure-api-key"

cd server

go run cmd/main.go