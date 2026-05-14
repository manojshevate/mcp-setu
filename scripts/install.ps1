#Requires -Version 5.0
<#
.SYNOPSIS
    mcp-setu installer for Windows
.DESCRIPTION
    Downloads and installs mcp-setu from GitHub releases with checksum verification
.PARAMETER Version
    Version to install (e.g., v0.1.2). If not specified, installs latest.
.PARAMETER InstallDir
    Installation directory. Defaults to $env:LOCALAPPDATA\mcp-setu\bin
.PARAMETER Force
    Skip confirmation and overwrite if already installed
.EXAMPLE
    irm https://raw.githubusercontent.com/manojshevate/mcp-setu/main/scripts/install.ps1 | iex
.EXAMPLE
    $env:MCP_SETU_VERSION='v0.1.2'; irm https://raw.githubusercontent.com/manojshevate/mcp-setu/main/scripts/install.ps1 | iex
#>

param(
    [string]$Version = $env:MCP_SETU_VERSION,
    [string]$InstallDir = $env:MCP_SETU_INSTALL_DIR,
    [switch]$Force = [bool]::Parse($env:MCP_SETU_FORCE ?? 'false')
)

# Enable strict error handling
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# Color helpers
$Colors = @{
    Red = "`e[91m"
    Green = "`e[92m"
    Plain = "`e[0m"
}

function Write-Status {
    param([string]$Message)
    Write-Host "${($Colors.Green)}>>>${($Colors.Plain)} $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "${($Colors.Red)}ERROR:${($Colors.Plain)} $Message" -ForegroundColor Red
    exit 1
}

function Write-Warning {
    param([string]$Message)
    Write-Host "${($Colors.Red)}WARNING:${($Colors.Plain)} $Message" -ForegroundColor Yellow
}

try {
    # Detect architecture
    $Arch = if ([Environment]::Is64BitOperatingSystem) {
        if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') {
            'arm64'
        } else {
            'amd64'
        }
    } else {
        Write-Error "Unsupported architecture: 32-bit Windows. Please use a 64-bit system."
    }

    Write-Status "Installing mcp-setu for Windows/$Arch"

    # Resolve version
    if (-not $Version) {
        Write-Status "Fetching latest release version..."
        try {
            $LatestRelease = Invoke-RestMethod -Uri 'https://api.github.com/repos/manojshevate/mcp-setu/releases/latest' -ErrorAction Stop
            $Version = $LatestRelease.tag_name
        } catch {
            $Version = 'v0.1.1'
            Write-Warning "Could not fetch latest version from GitHub API, using fallback: $Version"
        }
    }

    Write-Status "Installing mcp-setu $Version"

    # Resolve install directory
    if (-not $InstallDir) {
        $InstallDir = Join-Path $env:LOCALAPPDATA 'mcp-setu' 'bin'
    }

    # Ensure install directory exists
    if (-not (Test-Path $InstallDir)) {
        Write-Status "Creating installation directory: $InstallDir"
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    # Check if binary already exists
    $BinaryPath = Join-Path $InstallDir 'mcp-setu.exe'
    if (Test-Path $BinaryPath) {
        if (-not $Force) {
            if ($Host.UI.PromptForChoice('', "mcp-setu already exists at $BinaryPath. Overwrite?", @('&Yes', '&No'), 1) -eq 1) {
                Write-Error "Installation cancelled"
            }
        }
    }

    # Create temporary directory
    $TempDir = New-TemporaryDirectory
    try {
        # Download binary and checksums
        $ZipName = "mcp-setu_${Version}_windows_${Arch}.zip"
        $DownloadUrl = "https://github.com/manojshevate/mcp-setu/releases/download/${Version}/${ZipName}"
        $ChecksumsUrl = "https://github.com/manojshevate/mcp-setu/releases/download/${Version}/checksums.txt"

        $ZipPath = Join-Path $TempDir $ZipName
        $ChecksumsPath = Join-Path $TempDir 'checksums.txt'

        Write-Status "Downloading $ZipName..."
        try {
            Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -ErrorAction Stop
        } catch {
            Write-Error "Failed to download $DownloadUrl"
        }

        Write-Status "Downloading checksums..."
        try {
            Invoke-WebRequest -Uri $ChecksumsUrl -OutFile $ChecksumsPath -ErrorAction Stop
        } catch {
            Write-Error "Failed to download checksums from $ChecksumsUrl"
        }

        # Verify checksum
        Write-Status "Verifying checksum..."
        $ActualHash = (Get-FileHash -Path $ZipPath -Algorithm SHA256).Hash

        $ChecksumContent = Get-Content -Path $ChecksumsPath -Raw
        $ExpectedHashLine = $ChecksumContent | Select-String $ZipName | Select-Object -First 1
        if (-not $ExpectedHashLine) {
            Write-Error "Checksum not found for $ZipName in checksums.txt"
        }

        $ExpectedHash = ($ExpectedHashLine.ToString() -split '\s+')[0]

        if ($ActualHash -ne $ExpectedHash) {
            Write-Error "Checksum verification failed!`nExpected: $ExpectedHash`nActual:   $ActualHash"
        }

        Write-Status "$($Colors.Green)✓$($Colors.Plain) Checksum verified"

        # Extract binary
        Write-Status "Extracting binary..."
        Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

        # Find and move binary
        $ExtractedBinary = Get-ChildItem -Path $TempDir -Name 'mcp-setu.exe' -Recurse | Select-Object -First 1
        if (-not $ExtractedBinary) {
            Write-Error "Binary not found in extracted archive"
        }

        $ExtractedPath = Join-Path $TempDir $ExtractedBinary
        Move-Item -Path $ExtractedPath -Destination $BinaryPath -Force

        # Verify installation
        Write-Status "Verifying installation..."
        $VersionOutput = & $BinaryPath version 2>&1
        if (-not $VersionOutput) {
            Write-Error "Installation verification failed. Binary exists but version command failed."
        }

        Write-Status "$($Colors.Green)✓$($Colors.Plain) mcp-setu installed successfully at $BinaryPath"
        Write-Host ''

        # Check if install directory is in PATH
        $PathDirs = $env:PATH -split ';'
        if ($InstallDir -notin $PathDirs) {
            Write-Host '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'
            Write-Warning "$InstallDir is not in your PATH"
            Write-Host ''
            Write-Host "To add $InstallDir to your PATH, run:"
            Write-Host ''
            Write-Host "  `$env:PATH = `"$InstallDir;`$env:PATH`""
            Write-Host ''
            Write-Host "For permanent PATH modification, run:"
            Write-Host ''
            Write-Host "  [Environment]::SetEnvironmentVariable('PATH', `"$InstallDir;`$env:PATH`", 'User')"
            Write-Host ''
            Write-Host "Then restart PowerShell or your terminal."
            Write-Host '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'
        }

        Write-Host ''
        Write-Status "Run '$($Colors.Green)mcp-setu chat$($Colors.Plain)' to get started!"
        Write-Host ''

    } finally {
        # Cleanup temporary directory
        Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
    }

} catch {
    Write-Error $_.Exception.Message
}

# Helper function for compatibility with older PowerShell versions
if (-not (Get-Command New-TemporaryDirectory -ErrorAction SilentlyContinue)) {
    function New-TemporaryDirectory {
        $Parent = [System.IO.Path]::GetTempPath()
        $Name = [System.IO.Path]::GetRandomFileName()
        return New-Item -ItemType Directory -Path (Join-Path $Parent $Name)
    }
}
