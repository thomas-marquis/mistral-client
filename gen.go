package main_test

import _ "go.uber.org/mock/gomock"

//go:generate mockgen -package mocks -destination mocks/client.go github.com/thomas-marquis/mistral-client/mistral Client
//go:generate mockgen -package mocks -destination mocks/cache_engine.go github.com/thomas-marquis/mistral-client/mistral/internal/cache Engine
