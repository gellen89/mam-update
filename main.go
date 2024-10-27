package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gellen89/mam-update-go/internal/appdir"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	appdirs, err := appdir.GetAppDirs("mamdynupdate")
	if err != nil {
		panic(err)
	}
	err = appdirs.EnsureDirs()
	if err != nil {
		panic(err)
	}

	mamId := os.Getenv("MAM_ID")
	cookieFileExists, err := hasCookieFile(appdirs.Data, "MAM.cookie")
	if err != nil {
		panic(err)
	}
	if !cookieFileExists && mamId == "" {
		panic("no cookie file found, must provided MAM_ID for initialization")
	}

	currentIp, err := getCurrentIpAddress()
	if err != nil {
		panic(err)
	}

	if !cookieFileExists {
		resp, err := setIpAddress(&http.Client{}, logger)
		if err != nil {
			panic(err)
		}
		if !resp.Resp.Success {
			logger.Error(fmt.Sprintf("IP Address Update did not complete successfully. Message: %q", resp.Resp.Msg))
			return
		}
		logger.Debug(fmt.Sprintf("IP Address update completed successfully. Message: %q", resp.Resp.Msg))

		logger.Debug("writing cookies...")
		file, err := os.Create(filepath.Join(appdirs.Data, "MAM.cookie"))
		if err != nil {
			panic(err)
		}
		encoder := gob.NewEncoder(file)
		err = encoder.Encode(resp.Cookies)
		if err != nil {
			panic(err)
		}
		logger.Debug("successfully wrote cookies to file")
		err = os.WriteFile("last_run_time", []byte(time.Now().Format(time.RFC3339)), 0666)
		if err != nil {
			panic(err)
		}
		logger.Debug("IP Address update success")
		return
	}

	oldIpAddress, oldIpExists, err := getOldIpAddress(appdirs.Data, "MAM.ip")
	if err != nil {
		panic(err)
	}

	if oldIpExists && oldIpAddress == currentIp {
		logger.Info("IP Addresses have not changed, skipping...")
		return
	}
	logger.Info("IP Addresses have changed since previous run...")

	currentTime := time.Now()
	lastRunTime, hasLastRun, err := getLastRunTime(appdirs.Data, "last_run_time")
	if err != nil {
		panic(err)
	}
	if hasLastRun {
		duration := currentTime.Sub(*lastRunTime)
		if duration <= time.Hour {
			logger.Info("Last run was within the last hour, skipping...")
			return
		}
	}

	file, err := os.Open(filepath.Join(appdirs.Data, "MAM.cookie"))
	if err != nil {
		panic(err)
	}
	var cookies []*http.Cookie
	decoder := gob.NewDecoder(file)
	if err = decoder.Decode(&cookies); err != nil {
		panic(err)
	}
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		panic(err)
	}
	inputUrl, err := url.Parse(dynSeedBoxUrl)
	if err != nil {
		panic(err)
	}
	jar.SetCookies(inputUrl, cookies)
	client := &http.Client{Jar: jar}
	resp, err := setIpAddress(client, logger)
	if err != nil {
		panic(err)
	}

	if !resp.Resp.Success {
		logger.Error(fmt.Sprintf("IP Address Update did not complete successfully. Message: %q", resp.Resp.Msg))
		return
	}
	logger.Debug(fmt.Sprintf("IP Address update completed successfully. Message: %q", resp.Resp.Msg))

	logger.Debug("writing cookies...")
	file, err = os.Create(filepath.Join(appdirs.Data, "MAM.cookie"))
	if err != nil {
		panic(err)
	}
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(resp.Cookies)
	if err != nil {
		panic(err)
	}
	logger.Debug("successfully wrote cookies to file")
	err = os.WriteFile("last_run_time", []byte(time.Now().Format(time.RFC3339)), 0666)
	if err != nil {
		panic(err)
	}
	logger.Debug("IP Address update success")
}

const (
	dynSeedBoxUrl = "https://t.myanonamouse.net/json/dynamicSeedbox.php"
)

type dynamicSeedboxResponse struct {
	Success bool   `json:"Success"`
	Msg     string `json:"msg"`
}

type setIpAddressResponse struct {
	Resp    *dynamicSeedboxResponse
	Code    int
	Cookies []*http.Cookie
}

func setIpAddress(client *http.Client, logger *slog.Logger) (*setIpAddressResponse, error) {
	inputUrl, err := url.Parse(dynSeedBoxUrl)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(&http.Request{
		Method: "GET",
		URL:    inputUrl,
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	logger.Debug(string(body))
	if resp.StatusCode != 403 {
		return nil, fmt.Errorf("unable to update ip address %s: %s", resp.Status, string(body))
	}
	var jsonResp dynamicSeedboxResponse
	err = json.Unmarshal(body, &jsonResp)
	if err != nil {
		return nil, err
	}

	return &setIpAddressResponse{
		Resp:    &jsonResp,
		Code:    resp.StatusCode,
		Cookies: resp.Cookies(),
	}, nil
}

func hasCookieFile(dataDir string, fileName string) (bool, error) {
	_, err := os.Stat(filepath.Join(dataDir, fileName))
	if err != nil && !os.IsNotExist(err) {
		return false, err
	} else if err != nil && os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

func getLastRunTime(dataDir string, fileName string) (*time.Time, bool, error) {
	path := filepath.Join(dataDir, fileName)

	bits, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, false, err
	} else if err != nil && os.IsNotExist(err) {
		return nil, false, nil
	}
	output, err := time.Parse(time.RFC3339, string(bits))
	if err != nil {
		return nil, false, err
	}
	return &output, true, nil
}

func getOldIpAddress(dataDir string, fileName string) (string, bool, error) {
	path := filepath.Join(dataDir, fileName)

	bits, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", false, err
	} else if err != nil && os.IsNotExist(err) {
		return "", false, nil
	}
	return string(bits), true, nil
}

func getCurrentIpAddress() (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get("https://ifconfig.io/ip")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(ip)), nil
}
