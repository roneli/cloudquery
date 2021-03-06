package plugin

import (
	"fmt"
	"github.com/cloudquery/cloudquery/logging"
	"github.com/hashicorp/go-plugin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const defaultOrganization = "cloudquery"

type managedPlugin interface {
	Name() string
	Version() string
	Provider() CQProvider
	Close()
}

type remotePlugin struct {
	name     string
	version  string
	client   *plugin.Client
	provider CQProvider
}

// NewRemotePlugin creates a new remoted plugin using go_plugin
func newRemotePlugin(providerName, version string) (*remotePlugin, error) {
	pluginPath, _ := GetProviderPath(providerName, version)
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		VersionedPlugins: map[int]plugin.PluginSet{
			1: PluginMap,
		},
		Cmd:              exec.Command(pluginPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		SyncStderr:       os.Stderr,
		SyncStdout:       os.Stdout,
		Logger:           logging.NewZHcLog(&log.Logger, ""),
	})
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, err
	}
	raw, err := rpcClient.Dispense("provider")
	if err != nil {
		client.Kill()
		return nil, err
	}

	provider, ok := raw.(CQProvider)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("failed to cast plugin")
	}
	return &remotePlugin{
		name:     providerName,
		version:  version,
		client:   client,
		provider: provider,
	}, nil
}

func (r remotePlugin) Name() string { return r.name }

func (r remotePlugin) Version() string { return r.version }

func (r remotePlugin) Provider() CQProvider { return r.provider }

func (r remotePlugin) Close() {
	if r.client == nil {
		return
	}
	r.client.Kill()
}

type embeddedPlugin struct {
	name     string
	version  string
	provider CQProvider
}

// NewEmbeddedPlugin is a managed plugin that is created in-process, usually used for debugging purposes
func newEmbeddedPlugin(providerName, version string, p CQProvider) *embeddedPlugin {
	return &embeddedPlugin{
		name:     providerName,
		version:  version,
		provider: p,
	}
}

func (e embeddedPlugin) Name() string { return e.name }

func (e embeddedPlugin) Version() string { return e.version }

func (e embeddedPlugin) Provider() CQProvider { return e.provider }

func (e embeddedPlugin) Close() { return }


// GetProviderPath returns expected path of provider on file system from name and version of plugin
func GetProviderPath(name string, version string) (string, error) {
	org := defaultOrganization
	split := strings.Split(name, "/")
	if len(split) == 2 {
		org = split[0]
		name = split[1]
	}

	pluginDir := viper.GetString("plugin-dir")

	extension := ""
	if runtime.GOOS == "windows" {
		extension = ".exe"
	}
	return filepath.Join(pluginDir, ".cq", "providers", org, name, fmt.Sprintf("%s-%s-%s%s", version, runtime.GOOS, runtime.GOARCH, extension)), nil
}