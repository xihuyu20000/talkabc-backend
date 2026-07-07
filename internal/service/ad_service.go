package service

import (
	"backend/internal/model"
	"backend/internal/repository"
)

func GetLatestAdBanner() ([]model.AdBanner, error) {
	return repository.GetLatestAdBanner()
}

func GetAdList() ([]model.AdBanner, error) {
	return repository.GetLatestAdBanner()
}

func TrackAdClick(adID string) error {
	return nil
}

func TrackAdImpression(adID string) error {
	return nil
}