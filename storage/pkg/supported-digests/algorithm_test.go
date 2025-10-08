package supporteddigests

import (
	"sync"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTmpDigestForNewObjects(t *testing.T) {
	// Test that TmpDigestForNewObjects returns the default algorithm (SHA256)
	algorithm := TmpDigestForNewObjects()
	assert.Equal(t, digest.Canonical, algorithm)
	assert.Equal(t, "sha256", algorithm.String())
}

func TestTmpSetDigestForNewObjects(t *testing.T) {
	tests := []struct {
		name        string
		algorithm   digest.Algorithm
		expectError bool
		expected    digest.Algorithm
	}{
		{
			name:        "Set SHA256",
			algorithm:   digest.SHA256,
			expectError: false,
			expected:    digest.SHA256,
		},
		{
			name:        "Set SHA512",
			algorithm:   digest.SHA512,
			expectError: false,
			expected:    digest.SHA512,
		},
		{
			name:        "Set empty string (should default to SHA256)",
			algorithm:   "",
			expectError: false,
			expected:    digest.Canonical, // SHA256
		},
		{
			name:        "Set unsupported algorithm SHA384",
			algorithm:   digest.SHA384,
			expectError: true,
			expected:    digest.Canonical, // Should remain unchanged (default)
		},
		{
			name:        "Set unsupported algorithm MD5",
			algorithm:   digest.Digest("md5:invalid").Algorithm(),
			expectError: true,
			expected:    digest.Canonical, // Should remain unchanged (default)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TmpSetDigestForNewObjects(tt.algorithm)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported digest algorithm")
				// Verify algorithm wasn't changed
				assert.Equal(t, tt.expected, TmpDigestForNewObjects())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, TmpDigestForNewObjects())
			}
		})
	}
}

func TestAlgorithmPersistence(t *testing.T) {
	// Test that algorithm changes persist across multiple calls
	err := TmpSetDigestForNewObjects(digest.SHA512)
	require.NoError(t, err)
	assert.Equal(t, digest.SHA512, TmpDigestForNewObjects())

	// Verify it's still SHA512 after another call
	assert.Equal(t, digest.SHA512, TmpDigestForNewObjects())

	// Change to SHA256
	err = TmpSetDigestForNewObjects(digest.SHA256)
	require.NoError(t, err)
	assert.Equal(t, digest.SHA256, TmpDigestForNewObjects())

	// Verify it's still SHA256 after another call
	assert.Equal(t, digest.SHA256, TmpDigestForNewObjects())
}

func TestAlgorithmStringRepresentation(t *testing.T) {
	// Test SHA256 string representation
	err := TmpSetDigestForNewObjects(digest.SHA256)
	require.NoError(t, err)
	assert.Equal(t, "sha256", TmpDigestForNewObjects().String())

	// Test SHA512 string representation
	err = TmpSetDigestForNewObjects(digest.SHA512)
	require.NoError(t, err)
	assert.Equal(t, "sha512", TmpDigestForNewObjects().String())
}

func TestAlgorithmConcurrency(t *testing.T) {
	// Test concurrent reads and writes to ensure thread safety
	const numReaders = 10
	const numWriters = 10

	var wg sync.WaitGroup
	errCh := make(chan error, numWriters)
	readResults := make(chan digest.Algorithm, numReaders)

	// Start reader goroutines
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			algorithm := TmpDigestForNewObjects() // Read operation
			readResults <- algorithm
		}()
	}

	// Start writer goroutines - all writing the same algorithm
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := TmpSetDigestForNewObjects(digest.SHA512) // All writers set SHA512
			if err != nil {
				errCh <- err
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errCh)
	close(readResults)

	// Check for any errors
	for err := range errCh {
		assert.NoError(t, err)
	}

	// Verify all readers got a valid algorithm (either SHA256 or SHA512)
	for algorithm := range readResults {
		assert.Contains(t, []digest.Algorithm{digest.SHA256, digest.SHA512}, algorithm)
	}

	// Final check - should be SHA512 since all writers set it
	finalAlgorithm := TmpDigestForNewObjects()
	assert.Equal(t, digest.SHA512, finalAlgorithm)
}

func TestIsSupportedDigestAlgorithm(t *testing.T) {
	tests := []struct {
		name      string
		algorithm digest.Algorithm
		expected  bool
	}{
		{"SHA256", digest.SHA256, true},
		{"SHA512", digest.SHA512, true},
		{"Canonical", digest.Canonical, true},
		{"Empty string", "", true},
		{"SHA384", digest.SHA384, false},
		{"MD5", digest.Digest("md5:invalid").Algorithm(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSupportedDigestAlgorithm(tt.algorithm)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSupportedDigestAlgorithms(t *testing.T) {
	algorithms := GetSupportedDigestAlgorithms()
	expected := []digest.Algorithm{digest.SHA256, digest.SHA512}
	assert.Equal(t, expected, algorithms)
}

func TestGetDigestAlgorithmName(t *testing.T) {
	tests := []struct {
		name      string
		algorithm digest.Algorithm
		expected  string
	}{
		{"SHA256", digest.SHA256, "SHA256"},
		{"SHA512", digest.SHA512, "SHA512"},
		{"Canonical", digest.Canonical, "SHA256"}, // Canonical is SHA256
		{"Unknown", digest.Digest("unknown:invalid").Algorithm(), "unknown"},
		// Case-insensitive tests
		{"sha256 lowercase", digest.Algorithm("sha256"), "SHA256"},
		{"SHA256 uppercase", digest.Algorithm("SHA256"), "SHA256"},
		{"Sha256 mixed case", digest.Algorithm("Sha256"), "SHA256"},
		{"sHa256 mixed case", digest.Algorithm("sHa256"), "SHA256"},
		{"sha512 lowercase", digest.Algorithm("sha512"), "SHA512"},
		{"SHA512 uppercase", digest.Algorithm("SHA512"), "SHA512"},
		{"Sha512 mixed case", digest.Algorithm("Sha512"), "SHA512"},
		{"sHa512 mixed case", digest.Algorithm("sHa512"), "SHA512"},
		{"Unknown mixed case", digest.Algorithm("Unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDigestAlgorithmName(tt.algorithm)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDigestAlgorithmExpectedLength(t *testing.T) {
	tests := []struct {
		name           string
		algorithm      digest.Algorithm
		expectedLength int
		expectedFound  bool
	}{
		{"SHA256", digest.SHA256, 64, true},
		{"SHA512", digest.SHA512, 128, true},
		{"Canonical", digest.Canonical, 64, true}, // Canonical is SHA256
		{"Empty string", "", 0, false},
		{"SHA384", digest.SHA384, 0, false},
		{"MD5", digest.Digest("md5:invalid").Algorithm(), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, found := GetDigestAlgorithmExpectedLength(tt.algorithm)
			assert.Equal(t, tt.expectedLength, length)
			assert.Equal(t, tt.expectedFound, found)
		})
	}
}

func TestDetectDigestAlgorithmFromLength(t *testing.T) {
	tests := []struct {
		name          string
		length        int
		expectedAlg   digest.Algorithm
		expectedFound bool
	}{
		{"SHA256 length", 64, digest.SHA256, true},
		{"SHA512 length", 128, digest.SHA512, true},
		{"Invalid length 32", 32, digest.Algorithm(""), false},
		{"Invalid length 96", 96, digest.Algorithm(""), false},
		{"Invalid length 0", 0, digest.Algorithm(""), false},
		{"Invalid length -1", -1, digest.Algorithm(""), false},
		{"Invalid length 256", 256, digest.Algorithm(""), false},
		{"Edge case length 63", 63, digest.Algorithm(""), false},
		{"Edge case length 65", 65, digest.Algorithm(""), false},
		{"Edge case length 127", 127, digest.Algorithm(""), false},
		{"Edge case length 129", 129, digest.Algorithm(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			algorithm, found := DetectDigestAlgorithmFromLength(tt.length)
			assert.Equal(t, tt.expectedFound, found)
			if tt.expectedFound {
				assert.Equal(t, tt.expectedAlg, algorithm)
			} else {
				assert.Equal(t, digest.Algorithm(""), algorithm)
			}
		})
	}
}

func TestDetectDigestAlgorithmFromLengthConsistency(t *testing.T) {
	// Test that DetectDigestAlgorithmFromLength is consistent with GetDigestAlgorithmExpectedLength
	supportedAlgorithms := GetSupportedDigestAlgorithms()

	for _, algorithm := range supportedAlgorithms {
		expectedLength, supported := GetDigestAlgorithmExpectedLength(algorithm)
		if supported {
			detectedAlgorithm, found := DetectDigestAlgorithmFromLength(expectedLength)
			assert.True(t, found, "Should detect algorithm %s for length %d", algorithm.String(), expectedLength)
			assert.Equal(t, algorithm, detectedAlgorithm, "Detected algorithm should match expected algorithm")
		}
	}
}

func TestDetectDigestAlgorithmFromLengthAllSupportedLengths(t *testing.T) {
	// Test all supported lengths to ensure they are detected correctly
	expectedLengths := []int{64, 128} // SHA256 and SHA512 lengths

	for _, length := range expectedLengths {
		algorithm, found := DetectDigestAlgorithmFromLength(length)
		assert.True(t, found, "Should detect algorithm for length %d", length)
		assert.NotEqual(t, digest.Algorithm(""), algorithm, "Should return a valid algorithm for length %d", length)

		// Verify the detected algorithm has the expected length
		expectedLength, supported := GetDigestAlgorithmExpectedLength(algorithm)
		assert.True(t, supported, "Detected algorithm should be supported")
		assert.Equal(t, length, expectedLength, "Detected algorithm should have the expected length")
	}
}
