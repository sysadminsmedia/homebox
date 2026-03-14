package config

type NotifierConf struct {
	// AllowNets will allow specific networks through for generic notifiers.
	// If this is filled, only these networks will be allowed through.
	AllowNets []string `yaml:"allow_nets"`
	// BlockNets will block specific networks from generic notifiers.
	// If this is filled, these networks will be blocked from generic notifiers.
	BlockNets []string `yaml:"block_nets"`
	// BlockLocalhost will prevent generic notifiers from sending to localhost.
	BlockLocalhost bool `yaml:"block_localhost" conf:"default:false"`
	// BlockLocalNets will prevent generic notifiers from sending to local networks. (RFC1918)
	BlockLocalNets bool `yaml:"block_local_nets" conf:"default:false"`
	// BlockBogonNets will prevent generic notifiers from sending to bogon networks. (Reserved IPs)
	BlockBogonNets bool `yaml:"block_bogon_nets" conf:"default:true"`
	// BlockCloudMetadata will prevent generic notifiers from sending to known cloud metadata IPs.
	BlockCloudMetadata bool `yaml:"block_cloud_metadata" conf:"default:true"`
}
