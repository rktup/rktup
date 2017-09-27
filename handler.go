package rktup

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
)

type Discovery struct {
	URLTemplate string `json:"url_template"`
}

type Pubkey struct {
	URL string `json:"url"`
}

type Manifest struct {
	Discovery Discovery `json:"discovery"`
	Pubkey    Pubkey    `json:"pubkey"`
}

type HTTPHandler struct {
	hostname    string
	githubToken string

	tmplIndex     *template.Template
	tmplDiscovery *template.Template
}

func NewHTTPHandler(hostname, githubToken string) (*HTTPHandler, error) {
	dataTmplIndex, err := Asset("index.html")
	if err != nil {
		return nil, err
	}
	dataTmplDiscovery, err := Asset("ac-discovery.html")
	if err != nil {
		return nil, err
	}
	tmplIndex, err := template.New("index.html").Parse(string(dataTmplIndex))
	if err != nil {
		return nil, err
	}
	tmplDiscovery, err := template.New("index.html").Parse(string(dataTmplDiscovery))
	if err != nil {
		return nil, err
	}
	return &HTTPHandler{
		hostname,
		githubToken,
		tmplIndex,
		tmplDiscovery,
	}, nil
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	acDiscovery := r.URL.Query().Get("ac-discovery")
	if acDiscovery == "1" {
		h.discovery(w, r)
		return
	}
	if r.URL.Path == "" || r.URL.Path == "/" || r.URL.Path == "/index.html" {
		data := struct {
			Version string
		}{
			Version,
		}
		h.tmplIndex.Execute(w, data)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (h *HTTPHandler) discovery(w http.ResponseWriter, r *http.Request) {
	owner, repo, subPath, err := splitPath(r.URL.Path)
	if err != nil {
		log.Printf("bad path %q: %v\n", r.URL.Path, err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	githubURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s/.rktup.json", owner, repo, subPath)
	req, err := http.NewRequest("GET", githubURL, nil)
	if err != nil {
		log.Printf("failed to create request for %q: %v", githubURL, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Accept", "application/vnd.github.VERSION.raw")
	if h.githubToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", h.githubToken))
	}
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("failed to query github: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		// all good
	case 404:
		http.Error(w, "Not found", http.StatusNotFound)
		return
	case 401:
		log.Printf("got unauthorized from github for %q", githubURL)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	default:
		log.Printf("unexpected http status from github for %q: %s %d", githubURL, http.StatusText(resp.StatusCode), resp.StatusCode)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read response body: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	manifestURL := path.Join(h.hostname, owner, repo, subPath)
	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		log.Printf("failed to unmarshal rktup manifest from %q: %v", githubURL, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if manifest.Discovery.URLTemplate == "" {
		log.Printf("maninfest discovery url template is empty for %q", githubURL)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	discovery := struct {
		Prefix   string
		Template string
		Pubkey   string
	}{
		manifestURL,
		manifest.Discovery.URLTemplate,
		manifest.Pubkey.URL,
	}
	h.tmplDiscovery.Execute(w, discovery)
}

func splitPath(fullPath string) (owner string, repo string, subPath string, err error) {
	parts := strings.SplitN(strings.Trim(fullPath, "/ "), "/", 3)
	switch {
	case len(parts) < 2:
		err = fmt.Errorf("not enough parts")
	case len(parts) == 2:
		owner, repo = parts[0], parts[1]
	default:
		owner, repo, subPath = parts[0], parts[1], parts[2]
	}
	return
}
