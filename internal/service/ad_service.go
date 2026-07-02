package service

import (
	"backend/internal/model"
	"backend/internal/repository"
)

func GetLatestAdBanner() ([]model.AdBanner, error) {
	return repository.GetLatestAdBanner()
}