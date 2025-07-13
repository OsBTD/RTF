package models

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Helper function to randomly select from a list
func pick(options []string) string {
	return options[rand.Intn(len(options))]
}

// AssignAvatar downloads a Dicebear "adventurer" avatar and saves it as a PNG.
func AssignAvatar(username, gender string) (string, error) {
	rand.Seed(time.Now().UnixNano())

	baseURL := "https://api.dicebear.com/9.x/adventurer/png"
	seed := url.QueryEscape(username)

	// Gender-based traits
	var hairOptions, accessoryOptions []string
	switch strings.ToLower(gender) {
	case "male":
		hairOptions = []string{"short01", "short02", "short03", "short04", "short05", "short06"}
		accessoryOptions = []string{"glasses", "glasses:variant01", "glasses:variant03"}
	case "female":
		hairOptions = []string{"long01", "long02", "long03", "long04", "long05", "long06"}
		accessoryOptions = []string{"earrings:variant01", "earrings:variant03", "glasses:variant02"}
	default:
		return "", errors.New("invalid gender 'male' or 'female'")
	}

	// Build query params
	params := url.Values{}
	params.Set("seed", seed)
	params.Set("size", "128")
	params.Set("hair", pick(hairOptions))
	params.Set("hairColor", pick([]string{"0e0e0e", "3eac2c", "6a4e35", "dba3be", "ab2a18"}))
	params.Set("eyes", pick([]string{"variant01", "variant02", "variant03", "variant04", "variant05"}))
	params.Set("eyebrows", pick([]string{"variant01", "variant02", "variant03", "variant04"}))
	params.Set("mouth", pick([]string{"variant01", "variant02", "variant03", "variant04"}))
	params.Set("skinColor", pick([]string{"9e5622", "763900", "ecad80", "f2d3b1"}))
	params.Set("features", pick([]string{"freckles", "blush", "birthmark"}))
	params.Set("featuresProbability", "100")

	if len(accessoryOptions) > 0 {
		acc := pick(accessoryOptions)
		if strings.HasPrefix(acc, "glasses") {
			params.Set("glasses", strings.TrimPrefix(acc, "glasses:"))
			params.Set("glassesProbability", "100")
		}
		if strings.HasPrefix(acc, "earrings") {
			params.Set("earrings", strings.TrimPrefix(acc, "earrings:"))
			params.Set("earringsProbability", "100")
		}
	}

	// Final URL
	avatarURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// HTTP request
	resp, err := http.Get(avatarURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch avatar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("avatar API status code: %d, message: %s", resp.StatusCode, string(body))
	}

	// Ensure target directory exists
	dir := "../frontend/public/images/avatars/"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Save image file
	filePath := filepath.Join(dir, username+".png")
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	// Return relative path
	return "/public/images/avatars/" + username + ".png", nil
}
