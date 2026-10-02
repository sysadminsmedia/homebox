package config

type NotifierConf struct {
	// The options below apply to every notifier whose URL carries a user-supplied
	// destination host (generic://, gotify://, ntfy://, smtp://, ...), not just
	// generic webhooks. See validate.ValidateNotifierURL.

	// AllowNets will allow specific networks through for notifiers.
	// If this is filled, only these networks will be allowed through.
	AllowNets []string `yaml:"allow_nets"`
	// BlockNets will block specific networks from notifiers.
	// If this is filled, these networks will be blocked from notifiers.
	BlockNets []string `yaml:"block_nets"`
	// BlockLocalhost will prevent notifiers from sending to localhost. On by
	// default: loopback is the Homebox host/container itself (IPv6 ::1 was
	// already blocked via BlockBogonNets), so this is rarely a real notifier.
	BlockLocalhost bool `yaml:"block_localhost" conf:"default:true"`
	// BlockLocalNets will prevent notifiers from sending to local networks. (RFC1918)
	// Off by default so self-hosted notifiers on the LAN (gotify, ntfy, ...) keep working.
	BlockLocalNets bool `yaml:"block_local_nets" conf:"default:false"`
	// BlockBogonNets will prevent notifiers from sending to bogon networks. (Reserved IPs)
	BlockBogonNets bool `yaml:"block_bogon_nets" conf:"default:true"`
	// BlockCloudMetadata will prevent notifiers from sending to known cloud metadata IPs.
	BlockCloudMetadata bool `yaml:"block_cloud_metadata" conf:"default:true"`
	// Dns64Nets lists the IPv6 prefixes used for DNS64/NAT64 translation (RFC 6052).
	// IPv6 addresses inside these prefixes carry an embedded IPv4 address, which is
	// extracted and checked against the allow/block rules above so DNS64 synthesis
	// cannot be used to bypass them. Defaults to the RFC 6052 well-known prefix and
	// the RFC 8215 local-use prefix; override if your network uses a custom prefix.
	Dns64Nets []string `yaml:"dns64_nets" conf:"default:64:ff9b::/96;64:ff9b:1::/48,env:NOTIFIER_DNS64_NETS,flag:notifier-dns64-nets"`
}
