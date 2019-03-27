package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/radu-matei/cnab-go/pkg/bundle"
	"github.com/radu-matei/cnab-to-oci/remotes"
	"github.com/docker/distribution/reference"
	"github.com/spf13/cobra"
)

type pushOptions struct {
	input              string
	targetRef          string
	insecureRegistries []string
	allowFallbacks     bool
}

func pushCmd() *cobra.Command {
	var opts pushOptions
	cmd := &cobra.Command{
		Use:   "push <bundle file> [options]",
		Short: "Fixes and pushes the bundle to an registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.input = args[0]
			if opts.targetRef == "" {
				return errors.New("--target flag must be set with a namespace ")
			}
			return runPush(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.targetRef, "target", "t", "", "reference where the bundle will be pushed")
	cmd.Flags().StringSliceVar(&opts.insecureRegistries, "insecure-registries", nil, "Use plain HTTP for those registries")
	cmd.Flags().BoolVar(&opts.allowFallbacks, "allow-fallbacks", true, "Enable automatic compatibility fallbacks for registries without support for custom media type, or OCI manifests")
	return cmd
}

func runPush(opts pushOptions) error {
	var b bundle.Bundle
	bundleJSON, err := ioutil.ReadFile(opts.input)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bundleJSON, &b); err != nil {
		return err
	}
	resolverConfig := createResolver(opts.insecureRegistries)
	ref, err := reference.ParseNormalizedNamed(opts.targetRef)
	if err != nil {
		return err
	}

	err = remotes.FixupBundle(context.Background(), &b, ref, resolverConfig, remotes.WithEventCallback(displayEvent))
	if err != nil {
		return err
	}
	d, err := remotes.Push(context.Background(), &b, ref, resolverConfig.Resolver, opts.allowFallbacks)
	if err != nil {
		return err
	}
	fmt.Printf("Pushed successfully, with digest %q\n", d.Digest)
	return nil
}
