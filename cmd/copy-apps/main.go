package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	HarborURL = "https://harbor.qa.chorus-tre.ch"
	Username  = ""
	Password  = ""
	Project   = "apps"
)

type Repository struct {
	Name string `json:"name"`
}

type Artifact struct {
	Type string `json:"type"`
	Tags []*Tag `json:"tags"`
}

type Tag struct {
	Name string `json:"name"`
}

func getAuthToken() (string, error) {
	url := fmt.Sprintf("%s/service/token?service=harbor-registry&scope=registry:catalog:read", HarborURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(Username, Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResponse struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.Token, nil
}

func getRepositories() ([]*Repository, error) {
	url := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories?page_size=100", HarborURL, Project)
	fmt.Println("url", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(Username, Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch repositories: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	//fmt.Println("body", string(body))

	var repositories []*Repository
	if err := json.Unmarshal(body, &repositories); err != nil {
		return nil, err
	}

	for _, repo := range repositories {
		if len(repo.Name) > 5 && repo.Name[:5] == "apps/" {
			repo.Name = repo.Name[5:]
		}
	}

	//fmt.Println("repositories", repositories)

	return repositories, nil
}

func getTags(repoName string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts?page_size=100", HarborURL, Project, repoName)
	//fmt.Println("url", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(Username, Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch tags for repository %s: %s", repoName, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var artifacts []*Artifact
	if err := json.Unmarshal(body, &artifacts); err != nil {
		return nil, err
	}

	var tags []string
	for _, artifact := range artifacts {
		for _, tag := range artifact.Tags {
			tags = append(tags, tag.Name)
		}
	}

	return tags, nil
}

func main() {
	HarborURL_ := os.Getenv("HARBOR_URL")
	if HarborURL_ != "" {
		HarborURL = HarborURL_
	}
	Project_ := os.Getenv("HARBOR_PROJECT")
	if Project_ != "" {
		Project = Project_
	}
	Username = os.Getenv("HARBOR_USERNAME")
	Password = os.Getenv("HARBOR_PASSWORD")

	repos, err := getRepositories()
	if err != nil {
		fmt.Printf("Error fetching repositories: %v\n", err)
		return
	}

	for _, repo := range repos {
		fmt.Printf("Repository: %s", repo.Name)
		tags, err := getTags(repo.Name)
		if err != nil {
			fmt.Printf("Error fetching tags for repository %s: %v\n", repo.Name, err)
			continue
		}
		for _, tag := range tags {
			fmt.Printf(", %s", tag)
		}
		fmt.Println()
	}
}
