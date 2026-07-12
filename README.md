<h1 align="center">p2pwn</h1>  
<h3 align="center">Dahua cameras security scanner via P2P</h3>  
<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26%2B-blue?style=flat&logo=go&logoColor=white" alt="Go"/>
  <img src="https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-green" alt="Go"/>
  <img src="https://img.shields.io/github/license/thebadinteger/p2pwn" alt="License"/>
</p>  

<div align="center"><img src="https://raw.githubusercontent.com/thebadinteger/p2pwn/refs/heads/main/preview.png" alt="Preview"></div>  

---

- [Features](#Features)
- [Installation](#Installation)
- [Usage](#Usage)
- [Configuration](#Configuration)
- [Credits](#Credits)
- [Disclaimer](#Disclaimer)

## Features:  
- Multithreaded  
- Protocol support: `CGI, SDK`  
- CVE and weak credential detection  
- Flexible configuration  
- Snapshots saving  
- Import XML generation for Dahua software  

## Installation:  
**Via Go:**  
```shell
go install github.com/thebadinteger/p2pwn@main
```  
**Build from source:**  
```shell
git clone https://github.com/thebadinteger/p2pwn.git
cd p2pwn
go build
```  
Or download the latest binary from the **[Releases page](https://github.com/thebadinteger/p2pwn/releases/latest)**.  

## Usage:  
```shell
./p2pwn -i [input] -o [output] -t [threads] (-c [config])
```  
```
[-i, --input] Input file or specific target(s)
Format:
XXXXXXXXXX > Prefix
XXXXXXXXXXYYYYY > S/N
[-o, --output] Output folder for results
Default > DD-MM-YYYY_HH-MM-SS
If doesn't exist, will be created
[-t, --threads] Number of threads for scanning
Default > 100
[-c, --config] Path to config file
Default > config.toml
[-?, -h, --help] Get general help
```  

**Input format:**  
S/N or Prefix
```
XXXXXXXXXX
XXXXXXXXXXYYYYY
```  
```
XXXXXXXXXX,XXXXXXXXXXYYYYY
```  
**Output format:**  
pwned.txt:
```
# [DD-MM-YYYY HH:MM:SS] Scan started | input > {input} | count > {count} | output > {output} | threads > {threads}
[{model}] Credentials: {login}:{password} | S/N: {serial} | Channels: {channels} | IP: {ip} > {method}

# [DD-MM-YYYY HH:MM:SS] Scan finished
```  
import_*.xml:  
*(Maximum 64 devices per file)*  
```
<?xml version="1.0" encoding="UTF-8"?>
<DeviceManager version="2.0">
        <Device name="{serial}" domain="{serial}" port="37777" username="{login}" password="{encryptedpass}" protocol="1" connect="19" />
</DeviceManager>
```  
snapshots/:  
*Special password characters are not displayed in the file name*  
```
{serial}_{login}_{password}_{model}.jpg
```  

## Configuration:  
- `config.toml`  

In the configuration, you can set up scanning and pentesting, and enable or disable check methods.  
**Default config:**  
```toml
[scan] # Scan configuration
timeout = 5000 # Connection timeout in milliseconds
retries = 3 # Number of retries on connect
generate = 1048576 # How many S/N to generate on prefix (1-1048576)
nurses = 200 # Number of workers for checking S/N (online/offline)

[pwn] # Usage of different protocols and methods
snapshot = true # Take snapshots
protocol.cgi = true # Web CGI protocol
protocol.sdk = true # 37777 SDK protocol
methods.brute = true # Credentials brute force
methods.cve-2021-33044 = true # CVE-2021-33044
methods.cve-2021-33045 = true # CVE-2021-33045
methods.cve-2024-39943 = true # CVE-2024-39943

[brute] # Credentials brute force configuration
credentials = [
  { login = "admin", password = "admin" },
  { login = "666666", password = "666666" },
  { login = "888888", password = "888888" },
  { login = "admin", password = "admin123" },
  { login = "default", password = "tluafed" },
]

[dummy] # Credentials for "dummy" account added via CVE
login = "p2pwn" # 5-32 alphanumeric characters
password = "p2password" # 8-32 alphanumeric characters
```  
Tips:  
- `nurses` - The higher, the faster the S/N check (online/offline), but more packets
- `snapshot` - Takes snapshot of the first channel only. Turning it off can speed up the scan 
- `protocol.cgi` - Checks CVEs and brute
- `protocol.sdk` - Can only brute force (Can be turned off to speed up the scan if not scanning NVRs)
- `methods.cve-2021-33044`, `methods.cve-2021-33045` - Extracts admin credentials
- `methods.cve-2024-39943` - Adds a dummy account

## Credits  
Made by [badinteger](https://github.com/thebadinteger) `[GPL v3.0 License]`  
Special thanks: **[THANKS.md](THANKS.md)**  

## Disclaimer  
**This software is intended for educational and authorized testing purposes only.  
The author is not responsible for your actions or any misuse and damage caused by the software.**
