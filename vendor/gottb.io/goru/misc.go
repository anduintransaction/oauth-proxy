package goru

import (
	"os"
	"strconv"
	"strings"
)

const (
	GORU_VERSION = "3.1.12"
)

func Version() string {
	return GORU_VERSION
}

func Require(version string) {
	if VersionCompare(Version(), version) >= 0 {
		return
	}
	ErrPrintf(ColorRed, "goru version %q is required, current version is %q\n", version, Version())
	os.Exit(1)
}

func VersionCompare(currentVersion, requiredVersion string) int {
	currentMajor, currentMinor, currentRevision := parseVersion(currentVersion)
	requiredMajor, requiredMinor, requiredRevision := parseVersion(requiredVersion)
	if currentMajor == requiredMajor && currentMinor == requiredMinor && currentRevision == requiredRevision {
		return 0
	}
	if currentMajor > requiredMajor ||
		(currentMajor == requiredMajor && currentMinor > requiredMinor) ||
		(currentMajor == requiredMajor && currentMinor == requiredMinor && currentRevision >= requiredRevision) {
		return 1
	}
	return -1
}

func parseVersion(version string) (int, int, int) {
	pieces := strings.Split(version, ".")
	if len(pieces) != 3 {
		return 0, 0, 0
	}
	major, _ := strconv.ParseInt(pieces[0], 10, 64)
	minor, _ := strconv.ParseInt(pieces[1], 10, 64)
	revision, _ := strconv.ParseInt(pieces[2], 10, 64)
	return int(major), int(minor), int(revision)
}

type RunMode string

const (
	DevMode        RunMode = "development"
	CheckMode      RunMode = "check"
	ProductionMode RunMode = "production"
)

func GetRunMode() RunMode {
	mode := os.Getenv("RUNMODE")
	switch mode {
	case "development":
		return DevMode
	case "check":
		return CheckMode
	default:
		return ProductionMode
	}
}
