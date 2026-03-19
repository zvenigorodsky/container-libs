//go:build !windows

package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"go.podman.io/common/pkg/json-proxy"
	"go.podman.io/image/v5/signature"
	"go.podman.io/image/v5/types"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 3 || os.Args[1] != "--sockfd" {
		return fmt.Errorf("usage: %s --sockfd <fd>", os.Args[0])
	}
	sockfd, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return fmt.Errorf("invalid sockfd: %v", err)
	}

	manager, err := jsonproxy.NewManager(
		jsonproxy.WithSystemContext(func() (*types.SystemContext, error) {
			return &types.SystemContext{}, nil
		}),
		jsonproxy.WithPolicyContext(func() (*signature.PolicyContext, error) {
			policy, err := signature.DefaultPolicy(nil)
			if err != nil {
				return nil, err
			}
			return signature.NewPolicyContext(policy)
		}),
	)
	if err != nil {
		return err
	}
	defer manager.Close()
	return manager.Serve(context.Background(), sockfd)
}
