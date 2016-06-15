package config

import (
	"flag"
	"os"
	"testing"
)

func TestSetFlagOverridesUnset(t *testing.T) {
	// setup new flagset
	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
	// test with no flags set
	givenLog := "log.log"
	wantLog := givenLog
	givenConfig := "dftConfig.json"
	wantConfig := givenConfig

	SetFlagOverrides(&givenLog, &givenConfig)

	if givenLog != wantLog {
		t.Errorf("expected log value to be unchanged, but it was %q", givenLog)
	}
	if givenConfig != wantConfig {
		t.Errorf("expected config value to be unchanged, but it was %q", givenLog)
	}
}

func TestSetFlagOverridesSet(t *testing.T) {
	// setup new flagset
	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
	// test with no flags set
	givenLog := "log.log"
	wantLog := "cli.log"
	givenConfig := "dftConfig.json"
	wantConfig := "cfg.json"

	os.Args = []string{"", "-log", wantLog, "-config", wantConfig}
	SetFlagOverrides(&givenLog, &givenConfig)

	if givenLog != wantLog {
		t.Errorf("expected log value to be %q, but it was %q", wantLog, givenLog)
	}
	if givenConfig != wantConfig {
		t.Errorf("expected config value to be %q, but it was %q", wantConfig, givenConfig)
	}
}
