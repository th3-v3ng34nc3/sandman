#Requires -RunAsAdministrator
<#
.SYNOPSIS
    Sandman installer for Windows.
.DESCRIPTION
    This script provides an interactive menu to install the Sandman tool and its
    necessary dependencies on a Windows system. It requires administrative
    privileges to install software and modify the system PATH.
#>
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

#region Helper Functions

# Simple color-coded output functions
function Write-Ok { param([string]$Message) Write-Host "  `u{2714}  $Message" -ForegroundColor Green }
function Write-Warn { param([string]$Message) Write-Host "  `u{26A0}  $Message" -ForegroundColor Yellow }
function Write-Info { param([string]$Message) Write-Host "  `u{2192}  $Message" -ForegroundColor Cyan }
function Write-Err { param([string]$Message) Write-Host "  `u{2716}  $Message" -ForegroundColor Red }
function Write-Header { param([string]$Message) Write-Host "`n$Message`n" -ForegroundColor Cyan -BackgroundColor DarkBlue }

# Checks if a command is available in the PATH
function Test-CommandExists {
    param([string]$Command)
    return [bool](Get-Command $Command -ErrorAction SilentlyContinue)
}

# Checks for a dependency and prints its status
function Test-Dependency {
    param(
        [string]$Command,
        [string]$Label
    )
    if (Test-CommandExists $Command) {
        $path = (Get-Command $Command).Source
        Write-Ok "$Label already installed ($path)"
        return $true
    }
    else {
        Write-Warn "$Label not found"
        return $false
    }
}

# Adds a directory to the system-wide PATH if it's not already there
function Add-ToPath {
    param([string]$Directory)

    $scope = [System.EnvironmentVariableTarget]::Machine
    $path = [System.Environment]::GetEnvironmentVariable('Path', $scope)

    if (-not ($path -split ';' -contains $Directory)) {
        Write-Info "Adding '$Directory' to the system PATH..."
        $newPath = "$path;$Directory"
        [System.Environment]::SetEnvironmentVariable('Path', $newPath, $scope)
        # Update the current process's PATH
        $env:Path = $newPath
        Write-Ok "'$Directory' has been added to the PATH. You may need to restart your terminal."
    } else {
        Write-Info "'$Directory' is already in the system PATH."
    }
}

# Updates the process's PATH environment variable from the machine and user stores.
function Update-Path {
    Write-Info "Updating PATH environment variable for this session..."
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::Machine) + ";" + [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::User)
}

# A robust function to download a file
function Invoke-Download {
    param(
        [string]$Uri,
        [string]$OutFile
    )
    try {
        # PowerShell 3.0+ is preferred for TLS 1.2+ support
        if ($PSVersionTable.PSVersion.Major -ge 3) {
            Invoke-WebRequest -Uri $Uri -OutFile $OutFile -UseBasicParsing
        }
        # PowerShell 2.0 fallback
        else {
            $webClient = New-Object System.Net.WebClient
            $webClient.DownloadFile($Uri, $OutFile)
        }
    } catch {
        Write-Err "Failed to download from '$Uri'. Error: $_"
        throw
    }
}

# A robust function to extract a zip archive
function Expand-ZipArchive {
    param(
        [string]$SourcePath,
        [string]$DestinationPath
    )
    try {
        # PowerShell 5.0+
        if ($PSVersionTable.PSVersion.Major -ge 5) {
            Expand-Archive -Path $SourcePath -DestinationPath $DestinationPath -Force
        }
        # Fallback for older versions using .NET
        else {
            Add-Type -AssemblyName System.IO.Compression.FileSystem
            [System.IO.Compression.ZipFile]::ExtractToDirectory($SourcePath, $DestinationPath)
        }
    } catch {
        Write-Err "Failed to extract '$SourcePath'. Error: $_"
        throw
    }
}

# Installs a tool using winget if available
function Install-WithWinget {
    param(
        [string]$WingetId,
        [string]$ToolName
    )
    if (-not (Test-CommandExists 'winget')) {
        Write-Warn "winget not found. Cannot install $ToolName automatically."
        Write-Info "Please install '$ToolName' manually or install winget from the Microsoft Store."
        return $false
    }

    Write-Info "Installing $ToolName via winget..."
    try {
        winget install --id $WingetId -e --accept-source-agreements --accept-package-agreements
        return $true
    } catch {
        Write-Err "winget failed to install $ToolName. Error: $_"
        return $false
    }
}

#endregion Helper Functions

#region Installers

function Install-Trivy {
    Write-Header "Installing Trivy"
    if (Test-Dependency 'trivy' 'Trivy') { return }

    $installDir = "C:\ProgramData\Sandman\bin"
    New-Item -Path $installDir -ItemType Directory -Force | Out-Null

    Write-Info "Finding latest Trivy release..."
    $response = Invoke-WebRequest -Uri "https://github.com/aquasecurity/trivy/releases/latest" -UseBasicParsing
    # Extract version tag (e.g., v0.50.1)
    $trivyTag = ($response.BaseResponse.ResponseUri.Segments[-1]).TrimEnd('/')
    $trivyVersion = $trivyTag.TrimStart('v')

    $zipUrl = "https://github.com/aquasecurity/trivy/releases/download/$trivyTag/trivy_$($trivyVersion)_Windows-64bit.zip"
    $zipPath = Join-Path $env:TEMP "trivy.zip"
    $extractPath = Join-Path $env:TEMP "trivy_extract_$([guid]::NewGuid())"

    Write-Info "Downloading $zipUrl"
    Invoke-Download -Uri $zipUrl -OutFile $zipPath

    Write-Info "Extracting Trivy..."
    Expand-ZipArchive -SourcePath $zipPath -DestinationPath $extractPath
    Move-Item -Path "$extractPath\trivy.exe" -Destination "$installDir\trivy.exe" -Force

    # Cleanup
    Remove-Item $zipPath -Force
    Remove-Item $extractPath -Recurse -Force

    Add-ToPath -Directory $installDir
    Write-Ok "Trivy installed -> $(& "$installDir\trivy.exe" --version | Select-Object -First 1)"
}

function Install-Opengrep {
    Write-Header "Installing Opengrep"
    if (Test-Dependency 'opengrep' 'Opengrep') { return }

    $installDir = "C:\ProgramData\Sandman\bin"
    $exePath = Join-Path $installDir "opengrep.exe"
    New-Item -Path $installDir -ItemType Directory -Force | Out-Null

    $url = "https://github.com/opengrep/opengrep/releases/latest/download/opengrep-windows-amd64.exe"
    Write-Info "Downloading $url"
    Invoke-Download -Uri $url -OutFile $exePath

    Add-ToPath -Directory $installDir
    Write-Ok "Opengrep installed -> $(& "$installDir\opengrep.exe" --version | Select-Object -First 1)"
}

function Install-ClamAV {
    Write-Header "Installing ClamAV"
    if (Test-Dependency 'clamscan' 'ClamAV') {
        Write-Info "Updating virus definitions..."
        try {
            if (Test-Path "C:\Program Files\ClamAV\freshclam.exe") {
                & "C:\Program Files\ClamAV\freshclam.exe"
            } else {
                freshclam
            }
        } catch {
            Write-Warn "Could not run freshclam. It might be running as a service. Definitions are likely up-to-date."
        }
        return
    }

    if (Install-WithWinget -WingetId 'ClamAV.ClamAV' -ToolName 'ClamAV') {
        Update-Path
        Write-Info "Updating virus definitions (this may take a minute)..."
        try {
            Start-Sleep -Seconds 5 # Give the service a moment to settle after install
            if (Test-Path "C:\Program Files\ClamAV\freshclam.exe") {
                & "C:\Program Files\ClamAV\freshclam.exe"
            } else {
                freshclam
            }
        } catch {
            Write-Warn "freshclam update failed. Run 'freshclam' manually before scanning."
        }
        if (Test-Path "C:\Program Files\ClamAV\clamscan.exe") {
            Write-Ok "ClamAV installed -> $(& "C:\Program Files\ClamAV\clamscan.exe" --version | Select-Object -First 1)"
        } else {
            Write-Ok "ClamAV installed -> $(clamscan --version | Select-Object -First 1)"
        }
    }
}

function Install-ZAP {
    Write-Header "Installing OWASP ZAP"
    if (Test-Dependency 'zap-baseline.py' 'ZAP (zap-baseline.py)') { return }

    # ZAP requires Java and Python
    Write-Info "Checking for Java and Python (ZAP runtime requirements)..."
    if (-not (Test-CommandExists 'java')) { Install-WithWinget -WingetId 'Oracle.JavaRuntimeEnvironment' -ToolName 'Java' } else { Write-Ok "Java is already installed." }
    if (-not (Test-CommandExists 'python')) { Install-WithWinget -WingetId 'Python.Python.3' -ToolName 'Python 3' } else { Write-Ok "Python is already installed." }

    # Manual install of ZAP Core
    Write-Info "Downloading ZAP release..."
    $response = Invoke-WebRequest -Uri "https://github.com/zaproxy/zaproxy/releases/latest" -UseBasicParsing
    $zapVersion = ($response.BaseResponse.ResponseUri.Segments[-1]).Substring(1)

    $zipUrl = "https://github.com/zaproxy/zaproxy/releases/download/v$($zapVersion)/ZAP_${zapVersion}_Core.zip"
    $zipPath = Join-Path $env:TEMP "zap.zip"
    $extractPath = "C:\Program Files\ZAP"

    Write-Info "Downloading $zipUrl"
    Invoke-Download -Uri $zipUrl -OutFile $zipPath

    Write-Info "Extracting to $extractPath"
    Expand-ZipArchive -SourcePath $zipPath -DestinationPath $extractPath
    Remove-Item $zipPath

    $zapDir = Join-Path $extractPath "ZAP_$($zapVersion)"
    Add-ToPath -Directory $zapDir

    Write-Ok "ZAP installed -> $zapDir"
}

function Install-Sandman {
    Write-Header "Installing Sandman CLI"
    if (Test-Dependency 'sandman' 'Sandman') { return }

    $installDir = "C:\ProgramData\Sandman\bin"
    New-Item -Path $installDir -ItemType Directory -Force | Out-Null

    $url = "https://github.com/th3-v3ng34nc3/sandman/releases/latest/download/sandman_windows_amd64.zip"
    $zipPath = Join-Path $env:TEMP "sandman.zip"
    $extractPath = Join-Path $env:TEMP "sandman_extract_$([guid]::NewGuid())"

    Write-Info "Downloading $url"
    Invoke-Download -Uri $url -OutFile $zipPath

    Write-Info "Extracting Sandman..."
    Expand-ZipArchive -SourcePath $zipPath -DestinationPath $extractPath
    Move-Item -Path "$extractPath\sandman.exe" -Destination "$installDir\sandman.exe" -Force

    # Cleanup
    Remove-Item $zipPath -Force
    Remove-Item $extractPath -Recurse -Force

    Add-ToPath -Directory $installDir
    Write-Ok "Sandman installed -> $(& "$installDir\sandman.exe" version | Select-Object -First 1)"
}

#endregion Installers

#region Menu & Main

function Show-Menu {
    Clear-Host
    Write-Host @"
`n
  ┌────────────────────────────────────────────────────┐
  │           🌙  Sandman — Dependency Setup           │
  └────────────────────────────────────────────────────┘
`n
  What do you want to scan?

  1) Container images                →  trivy
  2) Source code for secrets         →  trivy
  3) Source code (SAST)              →  opengrep
  4) Infrastructure as Code (IaC)    →  trivy
  5) OS / package vulnerabilities    →  trivy
  6) Files for malware / viruses     →  clamav
  7) Live web applications (DAST)    →  zap + python + java
  8) All underlying scan engines     →  all engines
  9) Sandman CLI                     →  sandman binary
  10) Everything                     →  sandman + all engines

  0) Exit
`n
"@
    $choice = Read-Host "  Your choice [0-10]"
    return $choice
}

$choice = Show-Menu

switch ($choice) {
    '1' { Install-Trivy }
    '2' { Install-Trivy }
    '3' { Install-Opengrep }
    '4' { Install-Trivy }
    '5' { Install-Trivy }
    '6' { Install-ClamAV }
    '7' { Install-ZAP }
    '8' {
        Install-Trivy
        Install-Opengrep
        Install-ClamAV
        Install-ZAP
    }
    '9' { Install-Sandman }
    '10' {
        Install-Sandman
        Install-Trivy
        Install-Opengrep
        Install-ClamAV
        Install-ZAP
    }
    '0' { Write-Host "`n  Bye."; return }
    default {
        Write-Err "Invalid choice: $choice"
        exit 1
    }
}

Write-Host ""
Write-Ok "Setup complete. Run 'sandman --help' to get started."
Write-Host ""

#endregion Menu & Main