package driver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.podman.io/storage/internal/driver"
)

func TestParseDriverOption(t *testing.T) {
	tests := []struct {
		name       string
		option     string
		wantDriver string
		wantKey    string
		wantValue  string
		wantErr    string
	}{
		{
			name:       "valid option with overlay driver",
			option:     "overlay.mountopt=nodev",
			wantDriver: "overlay",
			wantKey:    "mountopt",
			wantValue:  "nodev",
		},
		{
			name:       "empty driver string before the dot",
			option:     ".mountopt=nodev",
			wantDriver: "",
			wantKey:    "mountopt",
			wantValue:  "nodev",
		},
		{
			name:       "not dot in option",
			option:     "mountopt=nodev",
			wantDriver: "",
			wantKey:    "mountopt",
			wantValue:  "nodev",
		},
		{
			name:       "valid option with overlay2 driver",
			option:     "overlay2.size=10G",
			wantDriver: "overlay2",
			wantKey:    "size",
			wantValue:  "10G",
		},
		{
			name:       "uppercase driver and key are converted to lowercase",
			option:     "ZFS.SIZE=20G",
			wantDriver: "zfs",
			wantKey:    "size",
			wantValue:  "20G",
		},
		{
			name:       "spaces are trimmed",
			option:     " key= val ",
			wantDriver: "",
			wantKey:    "key",
			wantValue:  "val",
		},
		{
			name:    "invalid driver returns an error",
			option:  "magicfs.size=5G",
			wantErr: `unknown driver "magicfs" in option "magicfs.size=5G"`,
		},
		{
			name:    "invalid syntax",
			option:  "myopt",
			wantErr: "unable to parse key/value option: myopt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, key, val, err := driver.ParseDriverOption(tt.option)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantDriver, driver)
			assert.Equal(t, tt.wantKey, key)
			assert.Equal(t, tt.wantValue, val)
		})
	}
}
